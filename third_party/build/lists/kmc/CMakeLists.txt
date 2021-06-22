#kmc sub cmake
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/constants.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/libkmc_config.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/openssl_config.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/libsecurec_config.cmake)

check_define(PROJECT_KMC)
check_define(PROJECT_SDP)
check_define(PROJECT_KMC_SDP)
project ("${PROJECT_KMC_SDP}")


check_define(PROJECT_OPENSSL)
check_define(JOBCOUNT)
check_define(KMC_C_SOURCE_DIR)
check_define(OPENSSL_MINIMUM_REQUIRED_VERSION)


include(ExternalProject)
include(TestBigEndian)

#add fake target for top level cmake 
add_custom_target(${PROJECT_KMC_SDP})

get_target_property(KMC_OPENSSL_INC openssl_crypto_lib PUBLIC_HEADER_DIR)
get_target_property(LIBSECUREC_INC libsecurec_shared PUBLIC_HEADER_DIR)

set(KMC_SRCS_DIR 
    ${KMC_C_SOURCE_DIR}/src/cac/
    ${KMC_C_SOURCE_DIR}/src/common/
    ${KMC_C_SOURCE_DIR}/src/kmc/)

set(KMC_SRCS_DIR 
    ${KMC_SRCS_DIR}
    ${KMC_C_SOURCE_DIR}/src/mip/other/)

foreach(SRC IN LISTS KMC_SRCS_DIR)
    file(GLOB_RECURSE TMP_C ${SRC}/*.c)
    set(KMC_SRCS ${KMC_SRCS} ${TMP_C})
    file(GLOB_RECURSE TMP_H ${SRC}/*.h)
    set(KMC_HDRS ${KMC_HDRS} ${TMP_H})
endforeach()

set(SDP_SRCS_DIR
    ${KMC_C_SOURCE_DIR}/src/sdp/)

file(GLOB_RECURSE SDP_HDRS ${SDP_SRCS_DIR}/*.h)
file(GLOB_RECURSE SDP_SRCS ${SDP_SRCS_DIR}/*.c)


add_library(${PROJECT_KMC} ${BUILD_LIB_TYPE} ${KMC_SRCS} ${KMC_HDRS})
add_library(${PROJECT_SDP} ${BUILD_LIB_TYPE} ${SDP_SRCS} ${SDP_HDRS})

target_link_libraries(${PROJECT_KMC} openssl_crypto_lib libsecurec_shared)
target_link_libraries(${PROJECT_SDP} ${PROJECT_KMC})

add_dependencies(${PROJECT_KMC_SDP} ${PROJECT_KMC} ${PROJECT_SDP})
add_dependencies(${PROJECT_KMC} ${PROJECT_OPENSSL} ${PROJECT_HWSEC})
add_dependencies(${PROJECT_SDP} ${PROJECT_OPENSSL} ${PROJECT_HWSEC})

set_target_properties(${PROJECT_KMC} PROPERTIES 
    VERSION ${KMC_SDP_VERSION}
    SOVERSION ${KMC_SDP_VERSION_MAJOR}
)

set_target_properties(${PROJECT_SDP} PROPERTIES 
    VERSION ${KMC_SDP_VERSION}
    SOVERSION ${KMC_SDP_VERSION_MAJOR}
)

#fix kmc and sdp export symbols
if(WIN32 AND MSVC)
    set_target_properties(${PROJECT_KMC} PROPERTIES 
        LINK_FLAGS /DEF:${CMAKE_CURRENT_LIST_DIR}/win/kmcv3_export.def)
    set_target_properties(${PROJECT_SDP} PROPERTIES 
        LINK_FLAGS /DEF:${CMAKE_CURRENT_LIST_DIR}/win/sdpv3_export.def)
endif()


#for sdp
target_compile_options(${PROJECT_SDP}
    PUBLIC -DWSEC_COMPILE_SDP
)

#for all
foreach(ITEM IN LISTS PROJECT_KMC PROJECT_SDP)
    target_compile_options(${ITEM}
        PUBLIC -DWSEC_COMPILE_CAC_OPENSSL
        PUBLIC -DWSEC_USE_OPENSSL_110
        PUBLIC -DDWSEC_ERR_CODE_BASE=0x00000000 
        PUBLIC -DWSEC_COMPILE_MIP_OTHER
        PUBLIC -DWSEC_WRI_LOG_AUTO_END_WITH_CRLF
        )

    if(UNIX)
        #PRODUCT OPTION
        target_compile_options(${ITEM}
            PUBLIC -Werror -Wall
            PUBLIC -O2
            )
        #SECURITY_COMPILE addtion
        target_compile_options(${ITEM}
            PUBLIC -Wconversion -Wformat-security -Wextra --param ssp-buffer-size=4 -Wl,-Bsymbolic
            -Werror
            )

        if("${CMAKE_SIZEOF_VOID_P}" STREQUAL "8")
            target_compile_options(${ITEM}
                PUBLIC -D__LP64__
                )
        endif()

        TEST_BIG_ENDIAN(IS_BIG_ENDIAN)
        if(IS_BIG_ENDIAN)
            target_compile_options(${ITEM}
                PUBLIC -DWSEC_CPU_ENDIAN_MODE=1
                )
        else()
            target_compile_options(${ITEM}
                PUBLIC -DWSEC_CPU_ENDIAN_MODE=2
                )
        endif()

    elseif(WIN32 AND MSVC)
        #to be add
        if("${CMAKE_SIZEOF_VOID_P}" STREQUAL "8")
            target_compile_options(${ITEM}
                PUBLIC -D_WIN64
                )
        endif()

        #x86 default to little endian
        target_compile_options(${ITEM}
            PUBLIC -DWSEC_CPU_ENDIAN_MODE=2
            )

        #for scecure c
        target_compile_options(${ITEM}
            PUBLIC -DSECUREC_IS_DLL_LIBRARY
            PUBLIC -DSECUREC_ONLY_DECLARE_MEMSET
            )
    else()
        message(FATAL_ERROR "Not Supported")
    endif()

endforeach()

target_include_directories(${PROJECT_KMC} PRIVATE
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/include/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/cac/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/common/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/kmc/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/sdp/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/mip/>
    #fix to other for compatibility
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/mip/other>
    #openssl and securec
    $<BUILD_INTERFACE:${KMC_OPENSSL_INC}/>
    $<BUILD_INTERFACE:${LIBSECUREC_INC}/>
)

target_include_directories(${PROJECT_SDP} PRIVATE
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/include/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/cac/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/common/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/kmc/>
    $<BUILD_INTERFACE:${KMC_C_SOURCE_DIR}/src/sdp/>
    #openssl and securec
    $<BUILD_INTERFACE:${KMC_OPENSSL_INC}/>
    $<BUILD_INTERFACE:${LIBSECUREC_INC}/>
)
