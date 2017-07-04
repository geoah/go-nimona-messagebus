# Message Bus

The message bus allows peers in the nimona network to exchange signed messages
between them. Messages have a pretty basic structure, they consist of a creator,
an arbitrary type (`string`), an arbitrary payload (`[]byte`), and a PGP 
signature.
