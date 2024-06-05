# 🕷️ www-crawler
a lighting fast web ⚡ crawler, designed to crawl the entire internet.

# Setup and Build

To build from source, you will need to have Go installed. If you use the pre-built binaries, this step is not necessary.

```bash
# Individual steps
$ git clone https://github.com/NotReeceHarris/www-crawler
$ cd www-crawler
$ go build -o crawler ./src/.
$ ./crawler

# One command
$ git clone https://github.com/NotReeceHarris/www-crawler && cd www-crawler && go build -o crawler ./src/. && ./crawler
```

On Windows, you will need 64-bit GCC. Additionally, use the flag **`CGO_ENABLED=1`** when building.


# Database structure

```db
Table domains {
  id integer [primary key]
  domain TEXT
}

Table paths {
  id integer [primary key]
  domain integer
  path text
  secure bool
  httpCode text
  scanned bool
  onHold bool
}

Table links {
  id integer [primary key]
  parent integer
  child integer
}

Table emails {
  id integer [primary key]
  email integer
  path integer
}

Ref: paths.domain > domains.id
Ref: emails.path > paths.id
Ref: links.parent > paths.id
Ref: links.child > paths.id
```

https://dbdiagram.io/d/665ed3e4b65d9338797257df
