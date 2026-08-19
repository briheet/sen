// Compute helpers behind the example program.
pub fn fib(n: u32) u64 {
    if (n < 2) return n;
    return fib(n - 1) + fib(n - 2);
}

pub fn mix(n: u32) u64 {
    var sum: u64 = 0;
    var i: u32 = 0;
    while (i < n) : (i += 1) {
        sum +%= fib(i % 12) * 3;
    }
    return sum;
}
