module github.com/iancooper/Practical-Messaging-Go/lookup

go 1.23

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.11.1
	// The SQLite driver. Note that go.mod is a MODULE manifest, not a package one: unlike a
	// .csproj or a pom.xml, it cannot say "only localcopy may see this". So the boundary is
	// not here -- it is the import block of every file under model/, and the compiler that
	// will not let a package name something it did not import. Exercise 1 drew the RabbitMQ
	// boundary exactly the same way.
	github.com/mattn/go-sqlite3 v1.14.52
	github.com/rabbitmq/amqp091-go v1.14.0
)
