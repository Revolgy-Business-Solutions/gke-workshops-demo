<?php
$host = '127.0.0.1:3306';
//$user = getenv('MYSQL_USER');
//$pass = getenv('MYSQL_PASSWORD');
$username = "USERNAME";
$password = "PASSWORD";
$dbname   = "DATABASE";
 
// Create connection
$conn = new mysqli($host, $username, $password, $dbname);

// Check connection
if ($conn->connect_error) {
    die("Connection failed: " . $conn->connect_error);
}

$sql = "SELECT user_id, user_name FROM users";
$result = $conn->query($sql);

if ($result->num_rows > 0) {
    // output data of each row
    while($row = $result->fetch_assoc()) {
        echo "user_id: " . $row["user_id"]. " - Name: " . $row["user_name"]. "<br>";
    }
} else {
    echo "0 results";
}

$conn->close();
?>
