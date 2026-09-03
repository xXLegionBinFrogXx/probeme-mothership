cmake_minimum_required(VERSION 3.16)
project(probeme-fakeprovider C)

set(CMAKE_C_STANDARD 11)

find_package(PkgConfig REQUIRED)
pkg_check_modules(PME REQUIRED probeme)

add_library(probeme_fake SHARED fake.c)
target_include_directories(probeme_fake PRIVATE ${PME_INCLUDE_DIRS})

add_library(probeme_broken SHARED broken.c)

set_target_properties(probeme_fake probeme_broken PROPERTIES
    LIBRARY_OUTPUT_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/build
)
