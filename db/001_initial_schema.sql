CREATE TABLE Users (
  sessionID TEXT PRIMARY KEY
);

CREATE TABLE Tests (
  userSessionID TEXT PRIMARY KEY,
  currentQuestion INTEGER NOT NULL DEFAULT 1,
  totalNumberOfQuestions INTEGER NOT NULL DEFAULT 50,
  FOREIGN KEY (userSessionID) REFERENCES Users(sessionID) ON DELETE CASCADE,
  CHECK (totalNumberOfQuestions >= 1),
  CHECK (totalNumberOfQuestions <= 500),
  CHECK (currentQuestion <= totalNumberOfQuestions),
  CHECK (currentQuestion >= 1)
);

CREATE TABLE Words (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  chineseCharacters TEXT NOT NULL,
  pinyin TEXT NOT NULL,
  meaning TEXT
);
CREATE INDEX idx_word_pinyin on Words(pinyin);

CREATE TABLE Questions (
  wordID INTEGER,
  testID TEXT,
  questionNumber INTEGER NOT NULL,
  usersAnswer TEXT,
  FOREIGN KEY (wordID) REFERENCES Words(id) ON DELETE RESTRICT,
  FOREIGN KEY (testID) REFERENCES Tests(userSessionID) ON DELETE CASCADE,
  PRIMARY KEY(testID, questionNumber)
);

-- ensure that when creating a question, the question number is within the bounds of the total number of questions prescribed
CREATE TRIGGER check_question_number 
BEFORE INSERT ON Questions
FOR EACH ROW
BEGIN
  SELECT
  CASE
    WHEN NEW.questionNumber <= (SELECT totalNumberOfQuestions FROM Tests WHERE userSessionID = NEW.testID) AND NEW.questionNumber >= 1 THEN
      NULL
    ELSE
      RAISE (ABORT, "questionNumber exceeds the allowed range for this test.")
  END;
END;