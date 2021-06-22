#build spec for KMC Onetrack

cmake_minimum_required (VERSION 3.10.2)

project (KMC-Onetrack)

#cmake funcs
include(CheckLibraryExists)
include(ProcessorCount)

ProcessorCount(N)
set(JOBCOUNT 1)
if(NOT N EQUAL 0)
    set(JOBCOUNT ${N})
endif()

if(UNIX)
    set(ENV{MAKEFLAGS} "-j ${JOBCOUNT}")
elseif(WIN32 AND MSVC)
    include(${CMAKE_CURRENT_LIST_DIR}/cmake/msvc_compile.cmake)
endif()

include(${CMAKE_CURRENT_LIST_DIR}/cmake/check_cmd_config.cmake)
include(${CMAKE_CURRENT_LIST_DIR}/cmake/security_compile.cmake)
include(${CMAKE_CURRENT_LIST_DIR}/cmake/version_config.cmake)

#function
include(${CMAKE_CURRENT_LIST_DIR}/cmake/function.cmake)
#path config
include(${CMAKE_CURRENT_LIST_DIR}/cmake/path_config.cmake)

#init config for cmake 
if(WIN32 AND MSVC)
    set(CMAKE_BUILD_TYPE Release)  # 该操作在 Linux 会添加 DNDEBUG -O3 编译选项，与公司要求不符合，产品可根据需要选择
endif()
set(CMAKE_CXX_STANDARD 11)
set(CMAKE_C_STANDARD 90)

set(CMAKE_BINARY_DIR "${CMAKE_SOURCE_DIR}/.build")
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/lib")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

set(KMC_ALL_TOPLEVEL_CMAKE_DIR ${CMAKE_CURRENT_LIST_DIR})
set(KMC_ALL_BINARY_DIR ${CMAKE_BINARY_DIR})

#init output dir
file(MAKE_DIRECTORY ${CMAKE_RUNTIME_OUTPUT_DIRECTORY})
file(MAKE_DIRECTORY ${CMAKE_LIBRARY_OUTPUT_DIRECTORY})


#define the sub project for kmc onetrack
set(PROJECT_HWSEC     "securec")
set(PROJECT_KMC_SDP   "kmcsdp")
set(PROJECT_KMC       "kmc")
set(PROJECT_SDP       "sdp")
set(PROJECT_OPENSSL   "openssl")
set(PROJECT_KMC_EXT   "kmcext")

if (NOT BUILD_STATIC)
    set(BUILD_LIB_TYPE "SHARED")
else()
    set(CMAKE_C_ARCHIVE_CREATE "<CMAKE_AR> qcD <TARGET> <LINK_FLAGS> <OBJECTS>")
    set(CMAKE_C_ARCHIVE_APPEND "<CMAKE_AR> qD <TARGET> <LINK_FLAGS> <OBJECTS>")
    set(CMAKE_C_ARCHIVE_FINISH "<CMAKE_RANLIB> -D <TARGET>")
    set(BUILD_LIB_TYPE "STATIC")
    set(BUILD_TESTS NO)
endif()

if(KMC_ONLY)
    set(PROJECT_KMC_JAR "")
    set(PROJECT_KMC_JNI "")
    set(PROJECT_KMC_EXT "")
endif()

set(TARGET_KMC_RELEASE "kmc_release")

#set(PROJECT_KMC_ONTRACK_UT "kmc_tests")
#
# project vars
set(KMC_BIN_DIR        "${CMAKE_BINARY_DIR}/kmc")
set(HWSEC_C_BIN_DIR    "${CMAKE_BINARY_DIR}/securec")
set(OPENSSL_BIN_DIR    "${CMAKE_BINARY_DIR}/openssl")
set(KMC_EXT_BIN_DIR    "${CMAKE_BINARY_DIR}/kmc-java/kmcext")

#sub projects
add_subdirectory("${CMAKE_CURRENT_LIST_DIR}/lists/securec" "${HWSEC_C_BIN_DIR}")
add_subdirectory("${CMAKE_CURRENT_LIST_DIR}/lists/openssl" "${OPENSSL_BIN_DIR}")
add_subdirectory("${CMAKE_CURRENT_LIST_DIR}/lists/kmc"     "${KMC_BIN_DIR}")
add_subdirectory("${CMAKE_CURRENT_LIST_DIR}/lists/kmc-ext" "${KMC_EXT_BIN_DIR}")

##kmc build with openssl and it deps openssl and hwsec-c 
add_dependencies("${PROJECT_KMC_SDP}" "${PROJECT_OPENSSL}")
add_dependencies("${PROJECT_KMC_SDP}" "${PROJECT_HWSEC}")
add_dependencies("${PROJECT_KMC_EXT}" "${PROJECT_KMC_SDP}")

#release install step via make install
set(EXPORT_FROM_DIR_LIB "${CMAKE_BINARY_DIR}/lib")
set(EXPORT_TO_DIR "${CMAKE_SOURCE_DIR}/release")

install(DIRECTORY "${EXPORT_FROM_DIR_LIB}" DESTINATION "${EXPORT_TO_DIR}" 
    FILE_PERMISSIONS  OWNER_READ OWNER_EXECUTE
	FILES_MATCHING
	PATTERN "*.a*"
	PATTERN "*.so*"
	PATTERN "*.dll"
    PATTERN "*.lib"
)

add_custom_target("${TARGET_KMC_RELEASE}" ALL DEPENDS 
    "${PROJECT_KMC_EXT}" "${PROJECT_KMC_SDP}"
    COMMAND ${CMAKE_COMMAND} -P ${${PROJECT_NAME}_BINARY_DIR}/cmake_install.cmake
)

#tests setup
# passed via cmake command-line eg cmake . -DBUILD_TESTS=ON
if(BUILD_TESTS AND PROJECT_KMC_TESTS)
    add_subdirectory("${CMAKE_CURRENT_LIST_DIR}/lists/tests" ${KMC_TESTS_BIN_DIR})
    add_dependencies("${PROJECT_KMC_TESTS}" "${TARGET_KMC_RELEASE}")
endif()

#debug turn debug log for kmcjni and kmcext
if(VERBOSE_DEBUG AND PROJECT_KMC_EXT AND PROJECT_KMC_JNI)
    target_compile_definitions(${PROJECT_KMC_EXT} PRIVATE VERBOSE_DEBUG)
endif()

