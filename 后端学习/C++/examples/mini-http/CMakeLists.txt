cmake_minimum_required(VERSION 3.16)
project(mini_http LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(mini_http src/main.cpp)

if(WIN32)
    target_link_libraries(mini_http PRIVATE ws2_32)
endif()
