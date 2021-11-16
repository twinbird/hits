# hits

http communication test server.

## Install

You can get latest binary on [this link](https://github.com/twinbird/hits/releases/latest).

## Uninstall

Just delete download binary.

```
$ rm ./hits
```

## Quick Start

```
$ ./hits
```

Now you can check the communication of the API.

```
curl -X POST \
	-d "key=your key" \
	-d "secret=your secret" \
	-d "param="parameter" \
	http://localhost:8080/your-nice-api/v1/post
```

The log will be displayed as below.

```
Time:
        2021/11/16 23:59:02
URL:
        /your-nice-api/v1/post
Method:
        POST
Protocol:
        HTTP/1.1
Header:
        Content-Length:47
        Content-Type:application/x-www-form-urlencoded
        User-Agent:curl/7.68.0
        Accept:*/*
Body:
        key=your key&secret=your secret&param=parameter
Parameters:
        key:your key
        secret:your secret
        param:parameter
```

## Option Example

First, you can get more information by looking at the help.

```
$ hits -h
```

### Logging to file

```
$ hits -o your_log_file.txt
```

### Specify response status

```
# not found
$ hits -s 404 
```

### Specify response header

```
$ hits -H 'Content-Type: text/csv; charset=utf8'
```

### Specify response body

```
$ echo "It works!" | hits
```

or

```
$ hits -r "It works!"
```

or 

```
$ hits -f response.html
```

### Response JSON

'-j' is shorthand of '-H 'Content-Type: application/json'.

```
$ hits -j -r '{"prop":"value"}'
```

### Use Basic auth

```
$ hits -u "user" -P "password"
```

## Advanced

In more complicated situations you can use a configuration file.

### Generate routing file template

```
$ hits -g route.json
```

### Use routing file

```
$ hits route.json
```

