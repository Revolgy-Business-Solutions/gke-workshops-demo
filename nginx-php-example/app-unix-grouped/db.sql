CREATE TABLE users(
     user_id MEDIUMINT NOT NULL AUTO_INCREMENT,
     user_name CHAR(30) NOT NULL,
     PRIMARY KEY (user_id)
);

INSERT INTO users (user_name) VALUES
    ('kot'),('mab');
