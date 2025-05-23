# Microsoft Azure Data Explorer (Kusto) [![GoDoc](https://godoc.org/github.com/Azure/azure-kusto-go?status.svg)](https://pkg.go.dev/github.com/Azure/azure-kusto-go/azkustodata) [![GoDoc](https://godoc.org/github.com/Azure/azure-kusto-go?status.svg)](https://pkg.go.dev/github.com/Azure/azure-kusto-go/azkustoingest)

- [About Azure Data Explorer](https://azure.microsoft.com/en-us/services/data-explorer/)
- [Data Client documentation](https://godoc.org/github.com/Azure/azure-kusto-go/azkustodata)
- [Ingest Client documentation](https://godoc.org/github.com/Azure/azure-kusto-go/azkustoingest)

# Version 1.0.0-preview Released (BREAKING CHANGES)
Version 1.0.0-preview introduced a significant change to the package structure, aligning Azure-Kusto-Go with all other Kusto SDKs structure.
The original package, `github.com/Azure/azure-kusto-go` is no longer published.
Instead, there are two new packages:
- `github.com/Azure/azure-kusto-go/azkustodata` - for query and management commands.
- `github.com/Azure/azure-kusto-go/azkustoingest` - for interacting with the ingesting data.

For more information, see the [migration guide](MIGRATION.md) and [changelog](CHANGELOG.md)


## Intro
This is a data plane SDK (it is for interacting with Azure Data Explorer (Kusto) service). For the control plane (resource administration), go [here](https://github.com/Azure/azure-sdk-for-go/tree/master/services/kusto/mgmt).

Use `github.com/Azure/azure-kusto-go/azkustodata` in your application to:

- Query Kusto/Azure Data Explorer clusters for rows, optionally into structs.

Use `github.com/Azure/azure-kusto-go/azkustoingest` in your application to:
- Import data into Kusto from local file, Azure Blob Storage file, Stream, or an `io.Reader`.


Key links:
- [Source code](https://github.com/Azure/azure-kusto-go)
- [API Reference Documentation](https://pkg.go.dev/github.com/Azure/azure-kusto-go)
- [Product documentation](https://azure.microsoft.com/en-us/services/data-explorer/)
- [Samples](https://pkg.go.dev/github.com/Azure/azure-kusto-go#readme-examples)

## Key concepts

Azure Data Explorer is a fully managed, high-performance, big data analytics platform that makes it easy to analyze high volumes of data in near real time. The Azure Data Explorer toolbox gives you an end-to-end solution for data ingestion, query, visualization, and management.

An Azure Data Explorer (Kusto) [**cluster**](https://docs.microsoft.com/azure/event-hubs/event-hubs-features#namespace) can have multiple databases. Each database, in turn, contains [**tables**](https://docs.microsoft.com/azure/event-hubs/event-hubs-features#partitions) which store data.

Query Azure Data Explorer with the [Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/query/), an open-source language initially invented by the team. The language is simple to understand and learn, and highly productive. You can use simple operators and advanced analytics.

For more information about Azure Data Explorer (Kusto), its features, and relevant terminology can be found here: [link](https://learn.microsoft.com/en-us/azure/data-explorer/data-explorer-overview)


## Getting started

### Install the package

Install the Kusto/Azure Data Explorer client module for Go with `go get`:

```bash
go get github.com/Azure/azure-kusto-go
```

### Prerequisites

- Go, version 1.22 or higher
- An [Azure subscription](https://azure.microsoft.com/free/)
- An [Azure Data Explorer Cluster](https://learn.microsoft.com/en-us/azure/data-explorer/).
- An Azure Data Explorer Database. You can create a Database in your Azure Data Explorer Cluster using the [Azure Portal](https://learn.microsoft.com/en-us/azure/data-explorer/create-cluster-database-portal).

## Examples

Examples for various scenarios can be found on [pkg.go.dev](https://pkg.go.dev/github.com/Azure/azure-kusto-go#readme-examples) or in the example*_test.go files in our GitHub repo for [azure-kusto-go](https://github.com/Azure/azure-kusto-go/tree/master/kusto).

### Create the connection string

Azure Data Explorer (Kusto) connection strings are created using a connection string builder for an existing Azure Data Explorer (Kusto) cluster endpoint of the form `https://<cluster name>.<location>.kusto.windows.net`.

```go
kustoConnectionStringBuilder := azkustodata.NewConnectionStringBuilder(endpoint)
```

### Create and authenticate the client

Azure Data Explorer (Kusto) clients are created from a connection string and authenticated using a credential from the [Azure Identity package][azure_identity_pkg], like [DefaultAzureCredential][default_azure_credential].
You can also authenticate a client using a system- or user-assigned managed identity with Azure Active Directory (AAD) credentials.

#### Using the `DefaultAzureCredential`

```go
// kusto package is: github.com/Azure/azure-kusto-go/azkustodata

// Initialize a new kusto client using the default Azure credential
kustoConnectionString := kustoConnectionStringBuilder.WithDefaultAzureCredential()
client, err := azkustodata.New(kustoConnectionString)
if err != nil {
	panic("add error handling")
}
// Be sure to close the client when you're done. (Error handling omitted for brevity.)
defer client.Close()
```

#### Using the `az cli`

```go
kustoConnectionString := kustoConnectionStringBuilder.WithAzCli()
client, err := azkustodata.New(kustoConnectionString)
```

#### Using a system-assigned managed identity

```go
kustoConnectionString := kustoConnectionStringBuilder.WithSystemManagedIdentity()
client, err := azkustodata.New(kustoConnectionString)
```

#### Using a user-assigned managed identity

```go
kustoConnectionString := kustoConnectionStringBuilder.WithUserAssignedIdentityClientId(clientID)
client, err := azkustodata.New(kustoConnectionString)
```

#### Using a k8s workload identity

```go
kustoConnectionString := kustoConnectionStringBuilder.WithKubernetesWorkloadIdentity(appId, tokenFilePath, authorityID)
client, err := kusto.New(kustoConnectionString)
```

#### Using a bearer token

```go
kustoConnectionString := kustoConnectionStringBuilder.WithApplicationToken(appId, token)
client, err := azkustodata.New(kustoConnectionString)
```

#### Using an app id and secret

```go
kustoConnectionString := kustoConnectionStringBuilder.WithAadAppKey(clientID, clientSecret, tenantID)
client, err := azkustodata.New(kustoConnectionString)
```

#### Using an application certificate

```go
kustoConnectionString := kustoConnectionStringBuilder.WithAppCertificate(appId, certificate, thumbprint, sendCertChain, authorityID)
client, err := azkustodata.New(kustoConnectionString)
```

### Querying

#### Simple queries

* Work for all types of requests, including queries and management commands.
* Limited to queries that can be built using a string literal known at compile time.

The simplest queries can be built using `kql.New`:

```go
q := kql.New("systemNodes | project CollectionTime, NodeId")
```

Queries can only be built using a string literals known at compile time, and special methods for specific parts of the query.  
The reason for this is to discourage the use of string concatenation to build queries, which can lead to security vulnerabilities.

#### Queries with parameters

* Can re-use the same query with different parameters.
* Only work for queries, management commands are not supported.

It is recommended to use parameters for queries that contain user input.  
Management commands can not use parameters, and therefore should be built using the builder (see next section).

Parameters can be implicitly referenced in a query:

```go
q := kql.New("systemNodes | project CollectionTime, NodeId | where CollectionTime > startTime and NodeId == nodeIdValue")
```

Here, `startTime` and `nodeIdValue` are parameters that can be passed to the query.

To Pass the parameters values to the query, create `kql.Parameters`:

```
params :=  kql.NewParameters().AddDateTime("startTime", dt).AddInt("nodeIdValue", 1)
```

And then pass it to the `Query` method, as an option:
```go
results, err := client.Query(ctx, database, query, QueryParameters(params))
if err != nil {
    panic("add error handling")
}

// You can see the generated parameters using the ToDeclarationString() method:
fmt.Println(params.ToDeclarationString()) // declare query_parameters(startTime:datetime, nodeIdValue:int);

// You can then use the same query with different parameters:
params2 :=  kql.NewParameters().AddDateTime("startTime", dt).AddInt("nodeIdValue", 2)
dataset, err = client.Query(ctx, database, query, QueryParameters(params2))
```

#### Queries with inline parameters
* Works for queries and management commands.
* More involved building of queries, but allows for more flexibility.

Queries with runtime data can be built using `kql.New`.
The builder will only accept the correct types for each part of the query, and will escape any special characters in the data.

For example, here is a query that dynamically accepts values for the table name, and the comparison parameters for the columns:

```go
dt, _ := time.Parse(time.RFC3339Nano, "2020-03-04T14:05:01.3109965Z")
tableName := "system nodes"
value := 1

q := kql.New("")
            .AddTable(tableName)
            .AddLiteral(" | where CollectionTime == ").AddDateTime(dt)
            .AddLiteral(" and ")
            .AddLiteral("NodeId == ").AddInt(value)

// To view the query string, use the String() method:
fmt.Println(q.String())
// Output: ['system nodes'] | where CollectionTime == datetime(2020-03-04T14:05:01.3109965Z) and NodeId == int(1)
```

Building queries like this is useful for queries that are built from user input, or for queries that are built from a template, and are valid for management commands too.



#### Query For Rows

The kusto `table` package queries data into a ***table.Row** which can be printed or have the column data extracted.

```go
// Query our database table "systemNodes" for the CollectionTimes and the NodeIds.
dataset, err := client.IterativeQuery(ctx, "database", q)
if err != nil {
	panic("add error handling")
}
// Don't forget to close the dataset when you're done. 
defer dataset.Close()

primaryResult := <-dataset.Tables() // The first table in the dataset will be the primary results.

// Make sure to check for errors.
if primaryResult.Err() != nil {
    panic("add error handling")
}

for rowResult := range primaryResult.Table().Rows() {
	if rowResult.Err() != nil {
        panic("add error handling")
	}
	row := rowResult.Row()
	
    fmt.Println(row) // As a convenience, printing a *table.Row will output csv
	// or Access the columns directly
    fmt.Println(row.IntByName("EventId"))
    fmt.Println(row.StringByIndex(1))
}

// Alternatively, use the `Query` method to get all of the data at once.
dataset, err := client.Query(ctx, "database", q)
if err != nil {
    panic("add error handling")
}

for _, row := range dataset.Tables()[0].Rows() {
    fmt.Println(row) // As a convenience, printing a *table.Row will output csv
    // or Access the columns directly
    fmt.Println(row.IntByName("EventId"))
    fmt.Println(row.StringByIndex(1))
}

```

#### Query Into Structs

Users will often want to turn the returned data into Go structs that are easier to work with.  The ***table.Row** object
that is returned supports this via the `.ToStruct()` method.

```go
// NodeRec represents our Kusto data that will be returned.
type NodeRec struct {
	// ID is the table's NodeId. We use the field tag here to instruct our client to convert NodeId to ID.
	ID int64 `kusto:"NodeId"`
	// CollectionTime is Go representation of the Kusto datetime type.
	CollectionTime time.Time
}

dataset, err := client.IterativeQuery(ctx, "database", q)
if err != nil {
panic("add error handling")
}
// Don't forget to close the dataset when you're done. 
defer dataset.Close()

primaryResult := <-dataset.Tables() // The first table in the dataset will be the primary results.

// Make sure to check for errors.
if primaryResult.Err() != nil {
panic("add error handling")
}

for result := range query.ToStructsIterative[NodeRec](primaryResult.Table()) {
    if result.Err != nil {
        panic("add error handling")
    }
    node := result.Out
    fmt.Println(node.ID)
}	

// Or use the `Query` method to get all of the data at once.
dataset, err := client.Query(ctx, "database", q)
if err != nil {
    panic("add error handling")
}

// You can use the `ToStructs` method directly on the dataset, or on a specific table.
structs, err := query.ToStructs[NodeRec](dataset)
if err != nil {
    panic("add error handling")
}

for _, node := range structs {
    fmt.Println(node.ID)
}

```

### Ingestion

The `azkustoingest` package provides access to Kusto's ingestion service for importing data into Kusto. This requires
some prerequisite knowledge of acceptable data formats, mapping references, etc.

That documentation can be found [here](https://docs.microsoft.com/en-us/azure/kusto/management/data-ingestion/)

If ingesting data from memory, it is suggested that you stream the data in via `FromReader()` passing in the reader
from an `io.Pipe()`. The data will not begin ingestion until the writer closes.


#### Creating a queued ingestion client
There are a few types of ingestion clients:
* Queued Ingest - `azkustoingest.New()` - the default client, uses queues and batching to ingest data. Most reliable.
* Streaming Ingest - `azkustoingest.NewStreaming()` - Directly streams data into the engine. Fast, but is limited with size and can fail.
* Managed Streaming Ingest - `azkustoingest.NewManaged()` - Combines a streaming ingest client with a queued ingest client to provide a reliable ingestion method that is fast and can ingest large amounts of data.
  Managed Streaming will try to stream the data, and if it fails multiple times, it will fall back to a queued ingestion.

To create an ingestion client, pass a Connection String, and additional options. 
```go
// queued client
kustoConnectionString := azkustodata.NewConnectionStringBuilder("<cluster>").WithDefaultAzureCredential()

// Queued ingestion client
in, err := azkustoingest.New(kustoConnectionString)
if err != nil {
	panic("add error handling")
}

// Streaming ingestion client with default database and table
in, err := azkustoingest.NewStreaming(kustoConnectionString, azkustoingest.WithDefaultDatabase("database"), azkustoingest.WithDefaultTable("table"))

// Managed streaming ingest client
in, err := azkustoingest.NewManaged(kustoConnectionString, azkustoingest.WithDefaultDatabase("database"), azkustoingest.WithDefaultTable("table"))

// Be sure to close the ingestor when you're done. (Error handling omitted for brevity.)
defer in.Close()
```

Queued ingestion client requires the url of the ingestion endpoint, usually starting with `ingest-`, and for streaming ingestion it's the opposite. 

The SDK will infer this endpoint from the given url. In case this is not wanted, you can use an option to disable it:

```go
in, err := azkustoingest.New(kustoConnectionString, azkustoingest.WithoutEndpointCorrection())
// Similarly, you can use azkustoingest.WithCustomIngestConnectionString() to provide a different query and ingest endpoint to a managed streaming ingest client.
in, err := azkustoingest.NewManaged(kustoConnectionString, azkustoingest.WithCustomIngestConnectionString(azkustodata.NewConnectionStringBuilder("https://ingest-<cluster>").WithDefaultAzureCredential()))
```

#### Ingestion From a File

Ingesting a local file requires simply passing the path to the file to be ingested:

```go
if _, err := in.FromFile(ctx, "/path/to/a/local/file"); err != nil {
	panic("add error handling")
}
```

`FromFile()` will accept Unix path names on Unix platforms and Windows path names on Windows platforms.
The file will not be deleted after upload (there is an option that will allow that though).

#### Ingestion From a Blob Storage File

This package will also accept ingestion from an Azure Blob Storage file:

```go
if _, err := in.FromFile(ctx, "https://myaccount.blob.core.windows.net/$root/myblob"); err != nil {
	panic("add error handling")
}
```

This will ingest a file from Azure Blob Storage. We only support `https://` paths and your domain name may differ than what is here.

#### Ingestion from an io.Reader

Sometimes you want to ingest a stream of data that you have in memory without writing to disk.  You can do this simply by chunking the
data via an `io.Reader`.

```go
r, w := io.Pipe()

enc := json.NewEncoder(w)
go func() {
	defer w.Close()
	for _, data := range dataSet {
		if err := enc.Encode(data); err != nil {
			panic("add error handling")
		}
	}
}()

if _, err := in.FromReader(ctx, r); err != nil {
	panic("add error handling")
}
```

It is important to remember that `FromReader()` will terminate when it receives an `io.EOF` from the `io.Reader`.  Use `io.Readers` that won't
return `io.EOF` until the `io.Writer` is closed (such as `io.Pipe`).

## Best Practices
See the SDK [best practices guide](https://docs.microsoft.com/azure/data-explorer/kusto/api/netfx/kusto-ingest-best-practices), which though written for the .NET SDK, applies similarly here.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Looking for SDKs for other languages/platforms?

- [Node](https://github.com/azure/azure-kusto-node)
- [Java](https://github.com/azure/azure-kusto-java)
- [.NET](https://docs.microsoft.com/en-us/azure/kusto/api/netfx/about-the-sdk)
- [Python](https://github.com/Azure/azure-kusto-python)
- [Azure CLI](https://learn.microsoft.com/en-us/azure/data-explorer/create-cluster-database-cli)
- [PowerShell](https://learn.microsoft.com/en-us/azure/data-explorer/create-cluster-database-powershell)
- [Azure Resource Manager template](https://learn.microsoft.com/en-us/azure/data-explorer/create-cluster-database-resource-manager)
