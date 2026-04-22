use snid::SNID;

fn main() {
    // Generate a new SNID
    let id = SNID::new();
    println!("Generated ID: {}", id);

    // Format as wire string with atom
    let wire = id.to_wire("MAT");
    println!("Wire format: {}", wire);

    // Parse a wire string
    let (parsed, atom) = SNID::parse_wire(&wire).expect("Failed to parse");
    println!("Parsed ID: {} (atom: {})", parsed, atom);

    // Generate batch of IDs
    let batch = SNID::generate_batch(5);
    println!("Generated {} batch IDs", batch.len());
    for (i, id) in batch.iter().enumerate() {
        println!("  [{}] {}", i, id.to_wire("MAT"));
    }
}
