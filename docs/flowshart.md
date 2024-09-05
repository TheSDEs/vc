# Flowchart

## Upload document to datastore

```mermaid
    sequenceDiagram;
    authentic source->>datastore: POST /notification;
    datastore->>authentic source: 200/400 ;
```


## Fetch a credential

```mermaid
    sequenceDiagram;
    wallet->>satosa;
    satosa->>apigw;
    apigw->>issuer;
    issuer->>apigw;
    apigw->>satosa;
    satosa->>wallet;
```
