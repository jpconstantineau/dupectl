# Files

``` mermaid
stateDiagram-v2
    [*] --> Pending : Insert from Scan
    Pending --> Progressing 
    Progressing --> Synced
    Synced --> Scanning
    Scanning --> Synced

    Scanning --> Missing: Not Found
    
    Hashed --> MarkedForDeletion: OPS-KeepPrimary

    UI --> MarkedForDeletion: API-Mark4Deletion
    Folders --> MarkedForDeletion: OPS-Mark4Deletion

    MarkedForDeletion --> ReadyForDeletion: OPS-Deleting
    ReadyForDeletion --> Deleting : API-GetFile2Delete
    Deleting --> Deleted : Not Found

    DuplicateFinder --> MarkedForHashing : OPS-SameFileSize + CreationDate
    MarkedForHashing --> Hashing : API-GetFileToHash
    Hashing --> Hashed : API-HashUpdated

    Hashing --> Missing: Not Found
  
    
    Missing  --> [*] : Garbage Collection    
    Deleted --> [*] : Garbage Collection

``` 
