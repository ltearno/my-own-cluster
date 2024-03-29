use std::env;
use std::fs;
use std::io::{Read, Write};

fn process(input_fname: &str, output_fname: &str) -> Result<(), String> {
    let mut input_file =
        fs::File::open(input_fname).map_err(|err| format!("error opening input: {}", err))?;
    let mut contents = Vec::new();
    input_file
        .read_to_end(&mut contents)
        .map_err(|err| format!("read error: {}", err))?;

        println!("Coming from the input buffer: {:?}!", contents);

    let mut output_file = fs::File::create(output_fname)
        .map_err(|err| format!("error opening output '{}': {}", output_fname, err))?;
    output_file
        .write_all(&contents)
        .map_err(|err| format!("write error: {}", err))
}

fn main() {
    let args: Vec<String> = env::args().collect();
    let program = args[0].clone();

    if args.len() < 3 {
        eprintln!("{} <input_file> <output_file>", program);
        return;
    }

    if let Err(err) = process(&args[1], &args[2]) {
        eprintln!("FROM RUST ERROR {}", err)
    }

    println!("Hello {:5}!", "x");
}