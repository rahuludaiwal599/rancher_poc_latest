When `rdctl factory-reset` is launched from the UI, it writes its stdout into
`TMP/rdctl-stdout.txt`

where on linux `TMP` is usually `/tmp`,

on macOS it's given by `$TMPDIR`

and on Windows by `%TEMP%`(command shell) or `$env:TEMP`(powershell).


This is most useful during development. When the UI runs in debug mode, it spawns `rdctl factory-reset` with the `--verbose` option.

We can't write the output into the `logs` directory as `factory-reset` deletes it.
