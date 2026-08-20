use std::{env, hint::black_box, io, sync::Arc};
use tokio::{
    io::{AsyncReadExt, AsyncWriteExt},
    net::{TcpListener, TcpStream},
};

#[tokio::main]
async fn main() -> io::Result<()> {
    let address = env::args()
        .nth(1)
        .unwrap_or_else(|| "127.0.0.1:8081".into());
    let listener = Arc::new(TcpListener::bind(&address).await?);
    eprintln!("Rust example listening on http://{address}");
    loop {
        let (stream, _) = listener.accept().await?;
        tokio::spawn(handle(stream));
    }
}

async fn handle(mut stream: TcpStream) -> io::Result<()> {
    let mut request = [0; 1024];
    let size = stream.read(&mut request).await?;
    let path = request[..size]
        .split(|byte| *byte == b' ')
        .nth(1)
        .unwrap_or(b"/");
    let (status, body) = if path.starts_with(b"/work") {
        ("200 OK", workload().to_string())
    } else if path == b"/health" {
        ("204 No Content", String::new())
    } else {
        ("200 OK", "hello from Rust\n".into())
    };
    let response = format!(
        "HTTP/1.1 {status}\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{body}",
        body.len()
    );
    stream.write_all(response.as_bytes()).await
}

#[inline(never)]
fn workload() -> u64 {
    let mut value = 0x9e37_79b9_7f4a_7c15_u64;
    for index in 0..300_000 {
        value = value.rotate_left(7) ^ index;
        value = value.wrapping_mul(0xbf58_476d_1ce4_e5b9);
    }
    black_box(value)
}
