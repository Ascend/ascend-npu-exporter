#openssl sub cmake
include(ExternalProject)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/constants.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/openssl_config.cmake)
include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/security_compile.cmake)

check_define(PROJECT_OPENSSL)
check_define(JOBCOUNT)
check_define(LOCAL_OPENSSL)
check_define(KMC_OPENSSL_VERSION)

project("${PROJECT_OPENSSL}")

# if not use local openssl just return
if(LOCAL_OPENSSL)
    check_define(OPENSSL_BIN_DIR)
    check_define(OPENSSL_LOCAL_INSTALL_DIR)
    check_define(KMC_OPENSSL_SOURCE_DIR)
    set(LOCAL_OPENSSL_WORKING_DIR "${KMC_OPENSSL_SOURCE_DIR}")

    if(UNIX)
        set(OPENSSL_CONFIG_COMMAND "${LOCAL_OPENSSL_WORKING_DIR}/config")
        set(OPENSSL_BUILD_COMMAND ${MAKE_TOOL} ${MAKE_CONCURRENT_PARAM} ${JOBCOUNT})
        set(SECURITY_COMPILE_OPTIONS "-fstack-protector-all -Wl,-z,relro -Wl,-z,now -D_FORTIFY_SOURCE=2 -O2 -s")
        message(${OPENSSL_CONFIG_COMMAND})
        message(${OPENSSL_BUILD_COMMAND})
        message(${SECURITY_COMPILE_OPTIONS})
    elseif(WIN32 AND MSVC)
        set(OPENSSL_CONFIG_COMMAND ${PERL} "<SOURCE_DIR>/${SUBPATH}/Configure" ${OPENSSL_PLANTFORM} ${NO_ASM})
        set(OPENSSL_BUILD_COMMAND ${MAKE_TOOL})
        set(SECURITY_COMPILE_OPTIONS "")
    else()
        message(FATAL_ERROR "not supported platform")
    endif()

    ExternalProject_Add("${PROJECT_OPENSSL}"
        SOURCE_DIR ${LOCAL_OPENSSL_WORKING_DIR}
        BINARY_DIR ${OPENSSL_BIN_DIR}/build
        STAMP_DIR ${OPENSSL_BIN_DIR}/stamp
        CONFIGURE_COMMAND ${OPENSSL_CONFIG_COMMAND}
        no-weak-ssl-ciphers 
        no-unit-test
        no-tests
        no-external-tests
        ${SECURITY_COMPILE_OPTIONS}
        --prefix=${OPENSSL_LOCAL_INSTALL_DIR}
        BUILD_COMMAND ${OPENSSL_BUILD_COMMAND}
        INSTALL_COMMAND ${MAKE_TOOL} install_dev
        #disable logs for openssl config and make
        LOG_CONFIGURE ON
        LOG_BUILD ON
        LOG_INSTALL ON
        )
else()
    #add fake target
    add_custom_target(${PROJECT_OPENSSL})
endif()

#check configure for openssl 
add_dependencies("${PROJECT_OPENSSL}" openssl_ssl_lib openssl_crypto_lib)
get_target_property(crypto_shared_path openssl_crypto_lib IMPORTED_LOCATION)

file(TO_NATIVE_PATH "${CMAKE_LIBRARY_OUTPUT_DIRECTORY}"  output_path)

#copy our libs to libs
add_custom_command(TARGET "${PROJECT_OPENSSL}"
    POST_BUILD
    COMMAND cmake -DWKDIR=${CMAKE_LIBRARY_OUTPUT_DIRECTORY}
    -DFROM_SO_NAME=${crypto_shared_path}
    -DSO_VERSION=${KMC_OPENSSL_VERSION}
    -DTO_DIR=${output_path}
    -P ${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/openssl_post_build.cmake
    )

if(WIN32 AND MSVC)
    get_target_property(crypto_imported_lib openssl_crypto_lib IMPORTED_IMPLIB)
    add_custom_command(TARGET "${PROJECT_OPENSSL}"
        POST_BUILD
        COMMAND cmake -E copy ${crypto_imported_lib} ${output_path}
        )
endif()

