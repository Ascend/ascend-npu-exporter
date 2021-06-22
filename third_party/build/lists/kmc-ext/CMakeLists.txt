include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/libkmc_config.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/libkmc_ext_config.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/libsecurec_config.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/openssl_config.cmake)

check_define(PROJECT_KMC_EXT)
project ("${PROJECT_KMC_EXT}")

check_define(KMCEXT_C_SOURCE_DIR)
check_define(KMCEXT_VERSION)
check_define(KMCEXT_VERSION_MAJOR)


set(KMC_EXT_LIBSRCS ${KMCEXT_C_SOURCE_DIR}/src)
file(GLOB_RECURSE KMC_EXT_SRCS ${KMC_EXT_LIBSRCS}/*.c )
file(GLOB_RECURSE KMC_EXT_HDRS ${KMC_EXT_LIBSRCS}/include/*.h)

message(${KMCEXT_C_SOURCE_DIR})

if(WIN32)
    file(GLOB_RECURSE TO_REMOVE_FILES ${KMC_EXT_LIBSRCS}/*_unix.c)
    list(REMOVE_ITEM KMC_EXT_SRCS ${TO_REMOVE_FILES})
else()
    file(GLOB_RECURSE TO_REMOVE_FILES ${KMC_EXT_LIBSRCS}/*_windows.c)
    list(REMOVE_ITEM KMC_EXT_SRCS ${TO_REMOVE_FILES})
endif()

add_library("${PROJECT_KMC_EXT}" ${BUILD_LIB_TYPE} ${KMC_EXT_SRCS} ${KMC_EXT_HDRS})


get_target_property(LIBKMC_INCLUDE_DIR libkmc_shared PUBLIC_HEADER_DIR)
get_target_property(LIBSDP_INCLUDE_DIR libsdp_shared PUBLIC_HEADER_DIR)
get_target_property(LIBSECUREC_INC libsecurec_shared PUBLIC_HEADER_DIR)
get_target_property(KMC_EXT_OPENSSL_INC openssl_crypto_lib PUBLIC_HEADER_DIR)

target_include_directories(${PROJECT_KMC_EXT} PRIVATE
    $<BUILD_INTERFACE:${KMCEXT_C_SOURCE_DIR}/include/>
    $<BUILD_INTERFACE:${KMCEXT_C_SOURCE_DIR}/src/common/>
    $<BUILD_INTERFACE:${KMCEXT_C_SOURCE_DIR}/src/seccrypto/>
    $<BUILD_INTERFACE:${LIBKMC_INCLUDE_DIR}/>
    $<BUILD_INTERFACE:${LIBSDP_INCLUDE_DIR}/>
    $<BUILD_INTERFACE:${KMC_EXT_OPENSSL_INC}/>
    $<BUILD_INTERFACE:${LIBSECUREC_INC}/>
)


if(WIN32 AND MSVC)
    target_compile_options("${PROJECT_KMC_EXT}"
        #windows has own security functions
        PUBLIC -DSECUREC_IS_DLL_LIBRARY
        PUBLIC -DSECUREC_ONLY_DECLARE_MEMSET
        )
endif()

target_link_libraries("${PROJECT_KMC_EXT}" 
    libkmc_shared 
    libsdp_shared
    openssl_crypto_lib
    libsecurec_shared
)


set_target_properties("${PROJECT_KMC_EXT}" PROPERTIES 
    VERSION ${KMCEXT_VERSION}
    SOVERSION ${KMCEXT_VERSION_MAJOR}
    #set export symbol default to hidden
    CXX_VISIBILITY_PRESET hidden
    C_VISIBILITY_PRESET hidden
)


