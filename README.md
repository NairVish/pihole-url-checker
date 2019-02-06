# Pi-hole URL Checker

This is a simple command-line program that checks to see if a specified URL is present in any active Pi-hole blacklists.

## Install

Install this on the system running Pi-hole. Assuming Go is configured correctly on this system:

```
go get github.com/NairVish/pihole-url-checker
```

## Usage

```
usage: pihole-url-checker [-h|--help] -q|--query "<value>" [-r|--root
                          "<value>"]

                          Checks if the given URL/query is present in any
                          (active) Pi-hole blocklists.

Arguments:

  -h  --help   Print help information
  -q  --query  Query URL to search
  -r  --root   Pi-hole's root folder. Default: /etc/pihole
```
  
For example, to determine which blocklists `example.com` may be in (if there are any to begin with):

```
pihole-url-checker -q example.com
```

The program will return any exact matches as well as approximate matches that may or may not result in a block.

Depending on the permissions of Pi-hole's root folder, you may need to run this as `sudo`:

```
sudo pihole-url-checker -q example.com
```

If the correct `GOROOT` is not part of the PATH for the root user, then you may need to call the executable using an absolute or relative path:

```
sudo /home/username/go/bin/pihole-url-checker -q example.com
```

## TODOs

This program is very much a work-in-progress!

- [ ] Whitelist (whitelist.txt) identification.
- [ ] Wildcard/regex blacklist (regex.list) identification.
- [ ] Multiple queries in single search.
- [ ] Speeding up search. (Currently, slow for large lists.)
- [ ] Recognize comments that are at the end of a line and ignore them. (Currently, returned as approximate matches even if the actual domain is the same.)
