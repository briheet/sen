// Senbon sampler wrapper. Compiled as the target's main module; imports the
// user's program as "user" and runs it under a SIGPROF CPU sampler.
// Samples are written directly from the signal handler (write is async-signal-safe),
// so no collector thread competes for the signal.
// Build: zig build-exe -lc --dep user=user -Mroot=sampler.zig -Muser=<entry>
const std = @import("std");
const builtin = @import("builtin");
const user = @import("user");

const itimerval = extern struct {
    it_interval: extern struct { tv_sec: isize, tv_usec: isize },
    it_value: extern struct { tv_sec: isize, tv_usec: isize },
};
extern "c" fn setitimer(which: c_int, new_value: *const itimerval, old_value: ?*itimerval) c_int;
const ITIMER_PROF: c_int = 2;

const samples_file_env = "SENBON_SAMPLES_FILE";
const sample_interval_us = 1000;
const sample_interval_ns = sample_interval_us * 1000;
const max_frames = 32;
const arch = builtin.cpu.arch;

var samples_fd: c_int = -1;

export fn senbon_marker() callconv(.c) void {}

fn readWord(addr: usize) usize {
    return @intFromPtr(@as(*const usize, @ptrFromInt(addr)));
}

fn captureStack(ctx: std.debug.cpu_context.Native, pcs: *[max_frames]usize) usize {
    var pc = ctx.getPc();
    var fp = ctx.getFp();
    var count: usize = 0;
    while (count < max_frames and pc != 0 and pc != std.math.maxInt(usize)) {
        pcs[count] = pc;
        count += 1;
        if (fp == 0 or fp == std.math.maxInt(usize)) break;
        const return_offset: usize = switch (arch) {
            .aarch64, .x86_64 => 8, // saved return address above the frame pointer
            else => break,
        };
        const next_fp = readWord(fp);
        if (next_fp <= fp) break; // frame pointers must grow toward older frames
        pc = readWord(fp + return_offset);
        fp = next_fp;
    }
    return count;
}

fn handler(sig: std.posix.SIG, info: *const std.posix.siginfo_t, ctx_ptr: ?*const anyopaque) callconv(.c) void {
    _ = sig;
    _ = info;
    if (samples_fd < 0) return;
    const ctx = std.debug.cpu_context.fromPosixSignalContext(ctx_ptr) orelse return;
    var pcs: [max_frames]usize = undefined;
    const count = captureStack(ctx, &pcs);
    if (count == 0) return;

    var line: [max_frames * 18]u8 = undefined;
    var pos: usize = 0;
    for (0..count) |i| {
        pos += (std.fmt.bufPrint(line[pos..], " {x}", .{pcs[i]}) catch return).len;
    }
    line[pos] = '\n';
    _ = std.c.write(samples_fd, line[0 .. pos + 1].ptr, pos + 1);
}

fn startSampler() !void {
    const path = std.c.getenv(samples_file_env) orelse return error.NoSamplesFile;
    const path_z = std.mem.span(path);
    const fd = std.c.open(path_z.ptr, .{ .ACCMODE = .WRONLY, .CREAT = true, .TRUNC = true }, @as(c_uint, 0o600));
    if (fd < 0) return error.OpenFailed;
    samples_fd = fd;

    var header: [64]u8 = undefined;
    const base = @intFromPtr(&senbon_marker);
    var pos: usize = 0;
    pos += (std.fmt.bufPrint(header[pos..], "#base {x}\n#interval {d}\n", .{ base, sample_interval_ns }) catch return error.BufferTooSmall).len;
    _ = std.c.write(fd, header[0..pos].ptr, pos);

    var action = std.posix.Sigaction{
        .handler = .{ .sigaction = handler },
        .mask = std.posix.sigemptyset(),
        .flags = std.posix.SA.SIGINFO | std.posix.SA.RESTART,
    };
    std.posix.sigaction(std.posix.SIG.PROF, &action, null);

    var timer = itimerval{
        .it_interval = .{ .tv_sec = 0, .tv_usec = sample_interval_us },
        .it_value = .{ .tv_sec = 0, .tv_usec = sample_interval_us },
    };
    _ = setitimer(ITIMER_PROF, &timer, null);
}

fn stopSampler() void {
    const timer = itimerval{
        .it_interval = .{ .tv_sec = 0, .tv_usec = 0 },
        .it_value = .{ .tv_sec = 0, .tv_usec = 0 },
    };
    _ = setitimer(ITIMER_PROF, &timer, null);
    if (samples_fd >= 0) _ = std.c.close(samples_fd);
    samples_fd = -1;
}

fn invoke(comptime main_fn: anytype) void {
    const info = @typeInfo(@TypeOf(main_fn));
    if (info != .@"fn") @compileError("senbon: expected a main function");
    if (info.@"fn".params.len != 0) @compileError("senbon: target main must take no arguments");
    const ret = info.@"fn".return_type orelse void;
    switch (@typeInfo(ret)) {
        .void, .noreturn => main_fn(),
        .error_union => |eu| switch (@typeInfo(eu.payload)) {
            void => main_fn() catch |err| reportError(err),
            else => _ = main_fn() catch |err| reportError(err),
        },
        else => _ = main_fn(),
    }
}

fn reportError(err: anyerror) void {
    std.debug.print("senbon: target main returned error: {s}\n", .{@errorName(err)});
}

pub fn main() void {
    startSampler() catch |err| {
        std.debug.print("senbon: sampler failed to start: {s}\n", .{@errorName(err)});
    };
    defer stopSampler();
    invoke(user.main);
}
