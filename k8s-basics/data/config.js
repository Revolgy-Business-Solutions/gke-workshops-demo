
var mysql = require('mysql');
var db_user = process.env.DB_USER
var db_pass = process.env.DB_PASSWORD

const config = {
  host: 'localhost',
  user: db_user,
  password: db_pass,
  database: 'demoapp'
};

const pool = mysql.createPool(config);

module.exports = pool;
