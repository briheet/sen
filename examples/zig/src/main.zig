// Senbon example: a compute pipeline to profile and visualize.
// Run under senbon: senbon run zig ./examples/zig
const std = @import("std");
const router = @import("router");

pub fn main() void {
    var total: u64 = 0;
    var batch: u32 = 1;
    while (batch <= 8) : (batch += 1) {
        total +%= router.run(batch * 2000);
    }
    std.debug.print("example result: {d}\n", .{total});
}
