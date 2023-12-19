# Bots Framework Database

The bot framework uses [dalgo](https://github.com/dal-go) (_Db Abstraction Layer in GO_) to work with different
databases in a unified way.

It was designed to be able to work with nested collections (e.g. with Firestore)
but thanks to `dalgo` it will be able to work with any database supported by `dalgo`.
Including relational SQL databases like MySQL, PostgreSQL, etc.

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