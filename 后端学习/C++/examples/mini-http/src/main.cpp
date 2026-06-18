// mini-http v2: 返回 HTTP 200 固定 HTML
// 详见 ../10-网络编程与简易HTTP服务.md
// Linux/WSL: cmake -S . -B build && cmake --build build && ./build/mini_http
// Windows: 需链接 ws2_32（CMakeLists 已配置）

#include <cstring>
#include <iostream>
#include <string>

#ifdef _WIN32
#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#endif
#include <winsock2.h>
#include <ws2tcpip.h>
#pragma comment(lib, "ws2_32.lib")
using socklen_t = int;
#define CLOSE_SOCKET closesocket
#else
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>
#define INVALID_SOCKET (-1)
#define SOCKET int
#define CLOSE_SOCKET close
#endif

namespace {

const char* kBody =
    "<html><body><h1>mini-http</h1><p>Hello from C++ demo-api</p></body></html>";

std::string build_response() {
    std::string body(kBody);
    return "HTTP/1.1 200 OK\r\n"
           "Content-Type: text/html; charset=utf-8\r\n"
           "Connection: close\r\n"
           "Content-Length: " +
           std::to_string(body.size()) + "\r\n\r\n" + body;
}

void init_sockets() {
#ifdef _WIN32
    WSADATA wsa;
    if (WSAStartup(MAKEWORD(2, 2), &wsa) != 0) {
        std::cerr << "WSAStartup failed\n";
        std::exit(1);
    }
#endif
}

void cleanup_sockets() {
#ifdef _WIN32
    WSACleanup();
#endif
}

}  // namespace

int main() {
    init_sockets();

    const int port = 8080;
    SOCKET server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd == INVALID_SOCKET) {
        std::cerr << "socket failed\n";
        return 1;
    }

    int opt = 1;
    setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR,
               reinterpret_cast<const char*>(&opt), sizeof(opt));

    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(static_cast<uint16_t>(port));

    if (bind(server_fd, reinterpret_cast<sockaddr*>(&addr), sizeof(addr)) < 0) {
        std::cerr << "bind failed (port " << port << " in use?)\n";
        CLOSE_SOCKET(server_fd);
        return 1;
    }

    if (listen(server_fd, 16) < 0) {
        std::cerr << "listen failed\n";
        CLOSE_SOCKET(server_fd);
        return 1;
    }

    std::cout << "mini-http listening on http://127.0.0.1:" << port << "/\n";

    const std::string response = build_response();

    while (true) {
        sockaddr_in client_addr{};
        socklen_t len = sizeof(client_addr);
        SOCKET client_fd =
            accept(server_fd, reinterpret_cast<sockaddr*>(&client_addr), &len);
        if (client_fd == INVALID_SOCKET) continue;

        char buf[1024];
        recv(client_fd, buf, sizeof(buf) - 1, 0);  // 简单忽略请求体

        send(client_fd, response.c_str(), static_cast<int>(response.size()), 0);
        CLOSE_SOCKET(client_fd);
    }

    CLOSE_SOCKET(server_fd);
    cleanup_sockets();
    return 0;
}
