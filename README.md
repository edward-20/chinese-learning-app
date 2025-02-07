# Deployment
Install sqlite3
```
cd db
sqlite3 chinese-learning-database.db
```
In SQLite:
```
.read 001_initial_schema.sql
.read 002_initialise_words.sql
```
Install go 1.23 or above
```
go build
./chinese-learning-app
```

# Set up Local Development Environment
Follow instructions to install Air on [https://github.com/air-verse/air](https://github.com/air-verse/air)

```
air
```
To watch TailwindCSS changes:
```
npm install
npx @tailwindcss/cli -i ./input.css -o ./static/index.css --watch
```
