# `cping`

`cping` is a simple ping utility that closely resembles that of a popular networking operating system. It supports:

- A CLI interface that should feel familiar to anyone who's used a network operating system - including the ability to abbreviate commands;
- IPv4 or IPv4 echo messages;
- Setting the length of the total packet (useful for finding MTU problems);
- Setting the number of echo messages sent to the destination;
- Setting the payload of the ICMP echo-response messages.

## Usage

See [usage.txt](./usage.txt) for up to date usage instructions (or run `cping` with `--help` or `help`).

## Examples

```
# ping 1.1.1.1 with default settings
cping 1.1.1.1

# ping 2606:4700:4700::1111 with default settings
cping ipv6 2606:4700:4700::1111

# ping 1.1.1.1 1000 times with 100-byte packets
cping size 100 count 1000 1.1.1.1

# the same command as above, but with abbreviations
cping si 100 co 1000 1.1.1.1
```

