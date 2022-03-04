# File Server

A software for file transfer through multiple clients using a server was developed using Golang as main programming language. If you want to run the software first of all you have to run *server.go* insider the directory server. After that you can run client.go and start to send requests to the server. Accepted commands:

    1 st: Stop client
    2 subscribe < channel >: Allows to subscribe to the channel admited
    3 send < file > < channel >: Allows to send a file through channel

Addiotionally you can modify the host and port in the server using flags like *port* and *host*.