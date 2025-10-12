# Harness 
A light TUI to explore and view all Kafka messages being sent with a given broker.

Built this to help visualize all of the black-box events that are sent to outside services at work.

## Usage
```bash
go install github.com/justinbather/harness@latest
```

```bash
harness <kafka broker host: defaults to localhost:9092>
```


### Keymaps
- j/k for up/down
- q or ctrl+c to quit
- d/u for up/down half page
- enter to view messages for a topic
- y to yank a messages body to your paste buffer


