# Folders
``` mermaid
stateDiagram-v2
    [*] --> Pending : Insert from Scan

    Pending --> Scanning: GetFolderToScanAPI

    Scanning --> Synced: No Folders/Files    
    Scanning --> Progressing : Folder/Files added

    Scanning --> Missing : Not found

    Progressing --> Synced : All Children Synced
    Progressing  --> Missing : Not found
    Synced --> Missing : Not found
    
    UI --> MarkedForDeletion: API-Mark4Deletion

    MarkedForDeletion --> ReadyForDeletion : No Children
    ReadyForDeletion --> Deleting : GetFolderDeleteAPI
    Deleting --> Deleted : Not found

    Missing  --> [*] : Garbage Collection    
    Deleted --> [*] : Garbage Collection



```
