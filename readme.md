# Flint

A bytecode virtual machine and compiler for a language that probably doesn't need to exist, but here we are.

## What is this?

Flint is a stack-based VM with its own instruction set, compiler, and all the trimmings. Think of it as "what if I built my own language runtime" but as a learning project instead of hubris.

## Features (that actually work)

- **Custom bytecode**: Because inventing your own opcodes is fun
- **Stack-based execution**: Push, pop, panic, repeat
- **Module system**: Import things, link things, pretend you're building a real language
- **Memory management**: Heap allocation that may or may not leak
- **Type system**: We've got integers of every size, floats, booleans, strings, and functions
- **Arithmetic**: Math works (mostly). Division by zero will panic you appropriately.
- **Syscalls**: Write to buffers! Feel like an OS!
- **Builtins**: Including the essential `panic` function for when things go wrong (they will)

## Architecture

```
Source → AST → Compiler → Bytecode → VM
```

- **common**: Core types (opcodes, constants, modules, the works)
- **compiler**: Turns AST into bytecode while crossing fingers
- **vm**: Executes bytecode using stacks, heaps, and prayer
- **ast**: Not shown but presumably exists

## Notable Design Decisions

- Two-byte operands because we're fancy
- Separate opcodes for signed/unsigned operations because type safety
- A `trap` instruction that does nothing but halt gracefully
- Module versioning (semantic versioning but make it bytes)
- The ability to load constants from other modules for that authentic dependency hell experience

## Running It

Look, if you've gotten this far, you probably know what `go test` does.

```bash
go test ./...
```

Watch the tests pass and feel accomplished.

## Why?

Learning, curiosity, and the unshakable belief that the world needs another toy VM.

## Status

Works on my machine. Your mileage may vary. PRs welcome but honestly this is mostly for educational purposes and mild amusement.

---

*"It compiles, it runs, and sometimes it even does what I expect."*
