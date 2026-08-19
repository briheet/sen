// Routes a workload batch through the services.
const services = @import("services");

pub fn run(batch: u32) u64 {
    var result: u64 = 0;
    var i: u32 = 0;
    while (i < batch) : (i += 1) {
        result +%= services.mix(1000);
    }
    return result;
}
