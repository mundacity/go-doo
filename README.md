
# Go-doo

## Description 

Go-doo is a simple command line notes/todo manager that lets you create, read & edit your notes or todo items. It'll support deletions soon. It uses SQLite for storage, but could easily support others. 

It uses a forgiving input parser so can handle spaces, meaning you're not required to use quotation marks like you are in commit messages, for example. So `godoo add -b input with spaces` would create an item with a body of `"input with spaces"`.

It also uses a shorthand date format, where `1y1m8d` is interpreted as 1 year, 1 month and 8 days from now. You can use full date strings like `2022-06-01` if you prefer. The date shorthand also supports negative numbers, so searching for an item with a deadline of `-8m` means the deadline was 8 months ago. 

You can also work with date ranges using this same shorthand. E.g. `godoo get -d -7d:7d` would return items with a deadline within a 14 day range, from 7 days before to 7 days from now. 

# Usage

## Creating new items

To create items, you use the `add` command with any combination of the following flags: 
| Flag | Name | Description | Example | Notes |
|------|------|-------------|---------|-------|
|-b | body | sets the body/content of the item you're creating | `add -b this is the body of the item` | `-b` can be omitted here and often elsewhere
|-c | childOf | the item will be the child of the item whose id is passed as the argument | `add -c 8` |
|-d | deadline | sets a deadline for the created item | `add -d 1 m 3 d` | same as `add -d 1m3d` |
|-t | tag | adds tag to created item | `add -t work` | item given 'work' tag | 

### Notes

Date ranges aren't supported for item creation. Multiple tag input is supported by `add -t t1*t2*t3`. The created item would have 't1', 't2' and 't3' tags. The delimiter (`*`) is configurable, but CLI interpreters don't allow certain characters, like semicolons.

You can often omit the body flag `-b` as the parser will try to figure out when the flag is  missing and where to add it in. This is to allow for quick item creation, basically getting your thoughts out quickly. Typically, if the body is the first thing you write then you won't need to explicitly state the `-b` flag, but if the body follows other flags with string-based arguments then you're more likely to run into problems. 

### Examples:

- `godoo add this item has no body flag` 
  - would create a new item with a body of "this item has no body flag"
- `godoo add flagless body -d 3d` 
  - would create a new item with a deadline 3 days from now, and a body of "flagless body"
- `godoo add -d 3d flagless body` 
  - would return an error because, when creating items, arguments for the deadline flag can be upto 10 characters (configurable) so the parser can't (yet) determine where one argument ends and the next one begins
- `godoo add -c 77 flagless body` 
  - would create an item with body of "flagless body" whose parent item's id is 77 because the parser is able to figure out where the argument for the `-c` flag (77) ends and can implicitly add the `-b` flag

You can use these flags in various combinations when creating items. As above, the `-b` flag will be assumed where possible:
- `godoo add add tests to project -t dev*testing d 1m`
  - would create an item with a body of "add tests to project", tags of "dev" & "testing", and a deadline one month from now.


## Retrieving items

To search for items that you have already created you use the `get` command, which supports the following flags:
| Flag | Name | Description | Example | Notes |
|------|------|-------------|---------|-------|
| -b | body | search by key phrase within body | `godoo get -b salmon fishcakes` | find items whose body contains phrase 'salmon fishcakes' |
| -i | id | search by id number | `godoo get -i 8` | get item with id of 8 |
| -d | deadline | search by deadline date | `godoo get -d 0d` | get items with a deadline of today |
| -e | creationDate | search by date item was created | `godoo get -e -7d:-3d` | get items created in a 4 day window between 7 and 3 days ago |
| -c | childOf | search by item's parent id | `godoo get -c 8` | get items with a parentId of 8 |
| -t | tag | search by tag | `godoo get -t dev`| return items marked with 'dev' tag |
| -a | all | get all items | `godoo get -a` | get every item |
| -f | finished | search by items marked as complete | `godoo get -f`| get all finished items |
| -F | unfinished | search by items marked as incomplete | `godoo get -F` | get all unfinished items|

### Notes

Whenever a new item is successfully created, the prompt outputs `Creation successful, ItemId: 3`. The number 3 here is just an example, the actual idNumber will obviously change each time. It's this id number that you can search for using the `-i` flag. 

You can use the flags listed in the above table in various combinations to build up very specific search criteria. The body flag can often be inferred in the same way described above, so it can be omitted in certain contexts.

### Examples

- `godoo get -F -d 0d`
  - find any unfinished (not done) items with a deadline of today
- `godoo get -f -d -8d`
  - find any complete/finished items with a deadline of 8 days ago
- `godoo get unique phrase -F -e -7d:0d`
  - find any unfinished items created in the last 7 days whose body contains the words 'unique phrase'
- `godoo get unique phrase -F -e -7d`
  - find any unfinished items created exactly 7 days ago whose body contains the words 'unique phrase'
- `godoo get -d -1m:1m -t notes -c 17 interesting fact -F -e -17d:0d`
  - find items with a deadline between 1 month ago and 1 month from now; with a tag called 'notes'; which is a child of item 17; whose body contains the phrase 'interesting fact'; that is not marked as finished/complete; and that was created some time over the last 17 days

Note that there is no sense checking for conflicting dates. So it is possible that the last example above could return one or more items with a deadline that pre-dates its creationDate. I think it's useful for retrospective note-taking but there are reasonable grounds for adding sense-checking in future. 

## Editing existing items

The command to use for changing or updating existing items is `edit`. There are two categories of flag for editing: lowercase for searching, or determining which item(s) to edit, and uppercase for actually modifiying the data of one or more items. 

This distinction enables you to use a single command to simultaneously search for and edit items. The searching component of this works in exactly the same way as it does for the `get` command (see above).  

| Flag | Function | Name | Description | Notes |
|------|------|---- |---------|---------|
| -b | search | body | search by key phrase in item body | |
| -i | search | id | search by item idNumber | |
| -d | search | deadline | search by item deadline | supports shorthand, longhand, date ranges |
| -c | search | childOf | search by parent idNumber | |
| -e | search | creationDate | search by creation date | supports shorthand, longhand, date ranges |
| -f | search | finished | search by completed items | |
| -B | edit | changeBody | edit the body field | |
| -C | edit | changeParent | change item's parent idNumber ||
| -D | edit | changeDeadline | change item's deadline | no date ranges |
| -F | edit | toggleComplete | toggle item's completion status | if complete, change to incomplete; if incomplete, change to complete|
| --append | behaviour | append | add new data to existing field | only relevant for string fields like item's body |
| --replace | behaviour | replace | replace existing data with new data |only relevant for string fields like item's body| 

### Notes

Date ranges are only supported by lowercase flags, or those with a 'search' function. Uppercase or editing flags do not support date ranges because a deadline is a specific date. 

### Examples

- `godoo edit -i 15 -F`
  - ex. 1: item is incomplete
    - item with idNumber 15 marked complete
  - ex. 2: item is complete
    - item with idNumber 15 marked incomplete
- `godoo edit -c 3 -F`
  - ex. 1: item incomplete
    - mark items with a parentId of 3 as done
  - ex. 2: item complete
    - mark items with a parentId of 3 as not done
- `godoo edit -b key phrase -D 1y`
  - find item/s with 'key phrase' in the body and change the deadline to 1 year from now
- `godoo edit -i 3 -B --append something interesting`
  - find item with id 3 and append the phrase 'something interesting' to the existing body
  - use the `--replace` flag to completely replace the body
    - if you don't pass either of them, you will be prompted to enter 'a' or 'r' 
  - placement of standalone flags like `--append`, `--replace`, `-f`, `-F` doesn't matter 
- `godoo edit -d -1m:5d -b golden badgers -e -3d:0d -B more common than you might think -D 12d --append`
  - find items that: have a deadline of between 1 month before today, and 5 days after today; whose bodies contain the phrase 'golden badgers'; and which were created at some point over the last 3 days
  - append the phrase 'more common than you might think' to the existing body, and change the deadline to 12 days from now

## Deleting items

Not yet supported but will be. 

## TODO

- Easier setup/installation
  - Configuration is via an `env` file (Viper). The path for this is currently set in the `config()` method in go-doo/cli/app.go
  - example config/env file in go-doo/example-env
- Support deletions 
  - need to decide if it should be by id only, or to allow the same search functionality as in the `get` and `edit` commands
- Add http support
  - For use on the same LAN, by individuals or small groups - don't see a use case for over the internet

