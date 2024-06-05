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

```
Table domains {
  id integer [primary key]
  domain TEXT
}

Table paths {
  id integer [primary key]
  domain integer
  path text
  secure bool
}

Table links {
  id integer [primary key]
  parent integer
  child integer
}

Ref: paths.domain > domains.id // many-to-one

Ref: links.parent > paths.id // many-to-one
Ref: links.child > paths.id // many-to-one
```

https://dbdiagram.io/d/665ed3e4b65d9338797257df

Public crawled database: [foo.db](./foo.db)
