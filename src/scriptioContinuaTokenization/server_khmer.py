from khmernltk import word_tokenize
import json
import socket 

def start_server():
    host = '127.0.0.1'
    port = 8080

    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.bind((host, port))
    server_socket.listen(5)

    print(f"Server listening on {host}:{port}")

    while True:
        client_socket, addr = server_socket.accept()
        print(f"Connection from {addr}")

        data = client_socket.recv(262144).decode('utf-8')
        

        print(word_tokenize(data))


        # Process the data or perform any desired actions

        responseObject = {
            "tokens":word_tokenize(data)
        }
        json_data = json.dumps(responseObject)
        print(json_data.encode("utf-8"))
        client_socket.send(json_data.encode("utf-8"))

        client_socket.close()

if __name__ == "__main__":
    start_server()
