cmake_minimum_required(VERSION 3.16)
project(hello_cmake LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_library(math_utils STATIC lib/math_utils.cpp)
target_include_directories(math_utils PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}/include)

add_executable(hello_cmake src/main.cpp)
target_link_libraries(hello_cmake PRIVATE math_utils)
