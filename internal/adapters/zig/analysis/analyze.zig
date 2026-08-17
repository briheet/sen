// Builds a static call graph and module map for a Zig project.
// Build with libc: zig run analyze.zig -lc -- <projectDir> <outputPath>
// Emits JSON:
// {root, entry, files: [{path, imports: [name], functions: [{id, name, startLine, startCol, endLine, endCol}]}], edges: [{from, to}]}
const std = @import("std");

const allocator = std.heap.page_allocator;

const FnDecl = struct {
    id: u32,
    name: []const u8,
    file_index: u32,
    start_line: usize,
    start_col: usize,
    end_line: usize,
    end_col: usize,
};

const Call = struct {
    callee: []const u8,
    enclosing: u32, // fn id, or maxInt(u32) if none
};

const FileInfo = struct {
    path: []const u8,
    imports: std.ArrayList([]const u8),
    functions: std.ArrayList(FnDecl),
    calls: std.ArrayList(Call),
};

const Edge = struct { from: u32, to: u32 };

var files_list = std.ArrayList(FileInfo).empty;
var edges = std.ArrayList(Edge).empty;
var next_id: u32 = 0;

const max_file_size = 1 << 20;

fn readFile(path: []const u8) ![]const u8 {
    const path_z = try allocator.dupeZ(u8, path);
    defer allocator.free(path_z);
    const fd = std.c.open(path_z, .{ .ACCMODE = .RDONLY });
    if (fd < 0) return error.OpenFailed;
    defer _ = std.c.close(fd);

    var buffer: [max_file_size]u8 = undefined;
    const n = std.c.read(fd, &buffer, buffer.len);
    if (n < 0) return error.ReadFailed;
    return allocator.dupe(u8, buffer[0..@intCast(n)]);
}

fn writeFile(path: []const u8, data: []const u8) !void {
    const path_z = try allocator.dupeZ(u8, path);
    defer allocator.free(path_z);
    const fd = std.c.open(path_z, .{ .ACCMODE = .WRONLY, .CREAT = true, .TRUNC = true }, @as(c_uint, 0o644));
    if (fd < 0) return error.OpenFailed;
    defer _ = std.c.close(fd);
    var written: usize = 0;
    while (written < data.len) {
        const n = std.c.write(fd, data[written..].ptr, data.len - written);
        if (n < 0) return error.WriteFailed;
        written += @intCast(n);
    }
}

fn collectFiles(dir: []const u8) !void {
    const dir_z = try allocator.dupeZ(u8, dir);
    defer allocator.free(dir_z);
    const dirp = std.c.opendir(dir_z) orelse return error.OpenFailed;
    defer _ = std.c.closedir(dirp);

    while (std.c.readdir(dirp)) |entry| {
        const name = std.mem.sliceTo(&entry.name, 0);
        if (std.mem.eql(u8, name, ".") or std.mem.eql(u8, name, "..")) continue;
        const full = try std.fs.path.join(allocator, &.{ dir, name });
        if (entry.type == std.c.DT.DIR) {
            try collectFiles(full);
            allocator.free(full);
        } else if (entry.type == std.c.DT.REG) {
            if (!std.mem.eql(u8, std.fs.path.extension(name), ".zig")) {
                allocator.free(full);
                continue;
            }
            if (std.mem.eql(u8, name, "build.zig") or std.mem.eql(u8, name, "build.zig.zon")) {
                allocator.free(full);
                continue;
            }
            try files_list.append(allocator, .{ .path = full, .imports = .empty, .functions = .empty, .calls = .empty });
        } else {
            allocator.free(full);
        }
    }
}

fn fnIndex(file_index: u32, id: u32) usize {
    for (files_list.items[file_index].functions.items, 0..) |decl, index| {
        if (decl.id == id) return index;
    }
    return 0;
}

fn analyzeFile(file_index: u32) !void {
    const file = &files_list.items[file_index];
    const source = try readFile(file.path);
    defer allocator.free(source);
    const text = try allocator.dupeZ(u8, source);
    defer allocator.free(text);

    var tokenizer = std.zig.Tokenizer.init(text);
    var prev_tag: std.zig.Token.Tag = .eof;
    var prev_text: []const u8 = "";
    var depth: usize = 0;
    var after_fn = false;
    var pending_name: []const u8 = "";
    var pending_start: usize = 0;
    var open_depth: usize = 0;
    var enclosing: u32 = std.math.maxInt(u32);

    while (true) {
        const token = tokenizer.next();
        if (token.tag == .eof) break;
        const token_text = text[token.loc.start..token.loc.end];

        if (after_fn and token.tag != .identifier) after_fn = false;
        switch (token.tag) {
            .keyword_fn => {
                after_fn = true;
                pending_name = "";
                pending_start = token.loc.start;
            },
            .identifier => {
                if (after_fn) {
                    pending_name = token_text;
                    after_fn = false;
                }
            },
            .l_brace => {
                if (pending_name.len != 0 and !std.mem.eql(u8, prev_text, ".")) {
                    const start_loc = std.zig.findLineColumn(text, pending_start);
                    const name = try allocator.dupe(u8, pending_name);
                    const id = next_id;
                    next_id += 1;
                    try file.functions.append(allocator, .{
                        .id = id,
                        .name = name,
                        .file_index = file_index,
                        .start_line = start_loc.line,
                        .start_col = start_loc.column,
                        .end_line = 0,
                        .end_col = 0,
                    });
                    pending_name = "";
                    enclosing = id;
                    open_depth = depth;
                }
                depth += 1;
            },
            .r_brace => {
                depth -= 1;
                if (enclosing != std.math.maxInt(u32) and depth == open_depth) {
                    const end_loc = std.zig.findLineColumn(text, token.loc.end);
                    const index = fnIndex(file_index, enclosing);
                    files_list.items[file_index].functions.items[index].end_line = end_loc.line;
                    files_list.items[file_index].functions.items[index].end_col = end_loc.column;
                    enclosing = std.math.maxInt(u32);
                }
            },
            .l_paren => {
                if (prev_tag == .identifier and !std.mem.eql(u8, prev_text, "_")) {
                    const callee = try allocator.dupe(u8, prev_text);
                    try file.calls.append(allocator, .{ .callee = callee, .enclosing = enclosing });
                }
            },
            .builtin => {
                if (std.mem.eql(u8, token_text, "@import")) {
                    const next1 = tokenizer.next();
                    const next2 = if (next1.tag == .l_paren) tokenizer.next() else null;
                    if (next2) |import_token| {
                        if (import_token.tag == .string_literal) {
                            const raw = text[import_token.loc.start..import_token.loc.end];
                            if (raw.len >= 2) {
                                const name = try allocator.dupe(u8, raw[1 .. raw.len - 1]);
                                try file.imports.append(allocator, name);
                            }
                        }
                    }
                }
            },
            else => {},
        }
        prev_tag = token.tag;
        prev_text = token_text;
    }
}

fn findFn(file_index: u32, name: []const u8) ?u32 {
    for (files_list.items[file_index].functions.items) |decl| {
        if (std.mem.eql(u8, decl.name, name)) return decl.id;
    }
    return null;
}

fn resolveFn(name: []const u8, file_index: u32) ?u32 {
    if (findFn(file_index, name)) |id| return id;
    for (0..files_list.items.len) |index| {
        if (index == file_index) continue;
        if (findFn(@intCast(index), name)) |id| return id;
    }
    return null;
}

fn callEnclosingFile(id: u32) u32 {
    for (files_list.items) |file| {
        for (file.functions.items) |decl| {
            if (decl.id == id) return decl.file_index;
        }
    }
    return 0;
}

fn rootAndEntry() !struct { root: u32, entry: []const u8 } {
    var root: u32 = std.math.maxInt(u32);
    var entry: []const u8 = "";
    for (files_list.items) |file| {
        for (file.functions.items) |decl| {
            if (std.mem.eql(u8, decl.name, "main")) {
                root = decl.id;
                entry = file.path;
            }
        }
    }
    if (root == std.math.maxInt(u32)) return error.NoMain;
    return .{ .root = root, .entry = entry };
}

fn addFmt(buffer: *std.ArrayList(u8), comptime format: []const u8, args: anytype) !void {
    const piece = try std.fmt.allocPrint(allocator, format, args);
    defer allocator.free(piece);
    try buffer.appendSlice(allocator, piece);
}

fn emitJson(buffer: *std.ArrayList(u8)) !void {
    const root_entry = try rootAndEntry();
    try addFmt(buffer, "{{\"root\":{d},\"entry\":\"{s}\",\"files\":[", .{ root_entry.root, root_entry.entry });
    for (files_list.items, 0..) |file, file_index| {
        if (file_index != 0) try buffer.appendSlice(allocator, ",");
        try addFmt(buffer, "{{\"path\":\"{s}\",\"imports\":[", .{file.path});
        for (file.imports.items, 0..) |import_name, index| {
            if (index != 0) try buffer.appendSlice(allocator, ",");
            try addFmt(buffer, "\"{s}\"", .{import_name});
        }
        try buffer.appendSlice(allocator, "],\"functions\":[");
        for (file.functions.items, 0..) |decl, index| {
            if (index != 0) try buffer.appendSlice(allocator, ",");
            try addFmt(buffer, "{{\"id\":{d},\"name\":\"{s}\",\"startLine\":{d},\"startCol\":{d},\"endLine\":{d},\"endCol\":{d}}}", .{
                decl.id, decl.name, decl.start_line, decl.start_col, decl.end_line, decl.end_col,
            });
        }
        try buffer.appendSlice(allocator, "]}");
    }
    try buffer.appendSlice(allocator, "],\"edges\":[");
    for (edges.items, 0..) |edge, index| {
        if (index != 0) try buffer.appendSlice(allocator, ",");
        try addFmt(buffer, "{{\"from\":{d},\"to\":{d}}}", .{ edge.from, edge.to });
    }
    try buffer.appendSlice(allocator, "]}");
}

pub fn main(init: std.process.Init) !void {
    var args = std.process.Args.Iterator.init(init.minimal.args);
    _ = args.next() orelse return error.InvalidArgs;
    const project_dir = args.next() orelse return error.InvalidArgs;
    const output_path = args.next() orelse return error.InvalidArgs;

    try collectFiles(project_dir);
    for (files_list.items, 0..) |_, file_index| {
        try analyzeFile(@intCast(file_index));
    }
    for (files_list.items) |file| {
        for (file.calls.items) |call| {
            if (call.enclosing == std.math.maxInt(u32)) continue;
            if (resolveFn(call.callee, callEnclosingFile(call.enclosing))) |target| {
                try edges.append(allocator, .{ .from = call.enclosing, .to = target });
            }
        }
    }
    _ = try rootAndEntry();

    var buffer = std.ArrayList(u8).empty;
    defer buffer.deinit(allocator);
    try emitJson(&buffer);
    try writeFile(output_path, buffer.items);
}
