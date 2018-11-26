# SqlGen
**This software is written for the article. [link](https://medium.com/@changhengliou/%E7%94%A8mysql-fulltext-search%E5%BB%BA%E7%AB%8B%E7%B0%A1%E6%98%93%E6%90%9C%E5%B0%8B%E5%BC%95%E6%93%8E-80659c28ec19)**

## To use this, please first execute the following sql.

```
CREATE DATABASE demo;
USE demo;
CREATE TABLE IF NOT EXISTS test (
  `id`    int not null auto_increment,
  `url`   varchar(255),
  `title` varchar(255),
  `text`  mediumtext,
  FULLTEXT `fulltext_index` (`title`, `text`),
  primary key (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4;
```

You can create init.sql and then use command line tool as follows.

`mysql -u [host] -P [port] -u"[user]" -p"[password]" < init.sql`

And then follow the instructions from the article.

**You can find the binary files [here](https://github.com/qq52184962/sqlgen/releases).** 

Or build your own, if you have golang sdk.

## Usage
`./sqlgen -h [host] -P [port] -p [password] -t [text_dir]`

**Example**

`./sqlgen -h localhost -P 3306 -p mypassword -t ./wikiextractor/text`
