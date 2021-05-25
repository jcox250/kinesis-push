# Kinesis Push

Kinesis push is a small tool for pushing data to a kinesis stream. It reads from stdin so you can either run the tool and enter the data you want to write to kinesis or you can put your data in a file and pipe it to the tool.

## Getting Started

Bring up localstack and create a kinesis stream that you can push data to
```
$ docker run --rm -it -p 4566:4566 -p 4571:4571 localstack/localstack
$ aws kinesis create-stream --stream-name test-stream --shard-count 5 --endpoint-url http://localhost:4566
```

test-data.txt
```
{"name": "foo", "age": 21}
{"name": "bar", "age": 22}
{"name": "fizz", "age": 23}
{"name": "buzz", "age": 24}
{"name": "foobar", "age": 25}
```

The -aws-region flag value doesn't matter when using localstack but it needs to be set to something
```
$ cat test-data.txt | ./kinesis-push -kinesis-endpoint http://localhost:4566 -stream-name test-stream -aws-region foo -debug
level=info msg="kinesis config" stream-name=test-stream role-arn= aws-region=foo endpoint=http://localhost:4566
level=info msg="service config" debug=true
level=debug msg="writing to kinesis..."
level=debug msg=put_record shard=shardId-000000000000
level=info msg="successfully wrote to kinesis" data="{\"name\": \"foo\", \"age\": 21}"
level=debug msg="writing to kinesis..."
level=debug msg=put_record shard=shardId-000000000001
level=info msg="successfully wrote to kinesis" data="{\"name\": \"bar\", \"age\": 22}"
level=debug msg="writing to kinesis..."
level=debug msg=put_record shard=shardId-000000000002
level=info msg="successfully wrote to kinesis" data="{\"name\": \"fizz\", \"age\": 23}"
level=debug msg="writing to kinesis..."
level=debug msg=put_record shard=shardId-000000000003
level=info msg="successfully wrote to kinesis" data="{\"name\": \"buzz\", \"age\": 24}"
level=debug msg="writing to kinesis..."
level=debug msg=put_record shard=shardId-000000000004
level=info msg="successfully wrote to kinesis" data="{\"name\": \"foobar\", \"age\": 25}"
```

## Potential improvements

Currently the script reads input from stdin and pushes to kinesis synchronously so if performance is a concern an improvement that could be made would be to read from stdin and push to kinesis asynchronously.
