# Telnet Compatibility and Control Characters

This server implements a plain text protocol over raw TCP.
It does not implement the Telnet protocol.

While tools like telnet may appear to work for basic testing, they can inject Telnet control sequences (for example when pressing Ctrl+C, Ctrl+D, or using special keys). These control bytes are not part of the server’s protocol and may cause the server’s input parser to block or desynchronize.

In particular:

- Ctrl+C in telnet does not send the characters ^ and c
- Telnet translates this into a Telnet control command (IAC + IP)
- The server treats these bytes as raw input and does not interpret them
- After receiving such bytes, the connection may no longer behave as expected

Because the server reads arbitrary byte chunks from the TCP stream and parses them directly as UTF-8 text, unexpected control bytes may leave the parser in an invalid state.