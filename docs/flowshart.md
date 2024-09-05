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
    wallet->>satosa: openID Federation;
    satosa->>apigw: POST /credential;
    apigw->>issuer: gRPC makeSDJWT();
    issuer->>apigw: Callback;
    apigw->>satosa: Callback;
    satosa->>wallet openID Federation;
```
