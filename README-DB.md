# Historical DALgo schema

`bots-fw` no longer depends on DALgo. This page records the key layout retained
by the optional `bots-fw-store-dalgo` adapter; it does not define a core API.
See [PERSISTENCE.md](PERSISTENCE.md) for migration and composition guidance.

## Database structure

- `botPlafforms` collection
    - `bots` collection
        - `botUsers` collection
        - `botChats` collection

In case if you use relational SQL databases, collections will be tables and documents will be rows.
The parent keys will be foreign key fields. You would not need `botPlafforms` & `bots` tables in this case.

- `botUsers` table
    - Platform string field
    - Bot string field
- `botChats` table
    - Platform string field
    - Bot string field

### botPlatforms collection

Contains the bot platforms like `telegram`, `whatsapp`, etc.

#### bots collection

Contains the bots. Each bot has a unique id.

##### botUsers collection

Contains the bot users. Each bot user has a unique id.
