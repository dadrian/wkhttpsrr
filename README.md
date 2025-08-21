# wkhttpsrr

This project implements the RFC Draft `draft-ietf-tls-wkech`. It is currently
based on draft revision 08. wkhttpsrr is a library that can take as input a hostname and parse an HTTPS RR configuration from .well-known, following the specification in the RFC draft. The data returned by the library can then be connected to any other program as needed.

This repository includes command-line utilties for fetching, validating, and printing an HTTPS RR configuration from a .well-known directory for an input hostname.

It does not (yet) include any integrations with an DNS software.
