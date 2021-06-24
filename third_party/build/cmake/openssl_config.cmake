
set(OPENSSL_MINIMUM_REQUIRED_VERSION 1.1.1)

#kmc openssl compat version
set(KMC_OPENSSL_MAJOR_VERSION 1)
set(KMC_OPENSSL_MINOR_VERSION 1)
if(UNIX)
    set(KMC_OPENSSL_VERSION ${KMC_OPENSSL_MAJOR_VERSION}.${KMC_OPENSSL_MINOR_VERSION})
elseif(WIN32 AND MSVC)
    set(KMC_OPENSSL_VERSION ${KMC_OPENSSL_MAJOR_VERSION}_${KMC_OPENSSL_MINOR_VERSION})
else()
    message(FATAL_ERROR "not support platform")
endif()


add_library(openssl_ssl_lib SHARED IMPORTED)
add_library(openssl_crypto_lib SHARED IMPORTED)

#pass via command line
if(USE_LOCAL_OPENSSL)
    set(LOCAL_OPENSSL TRUE)
else()
    set(LOCAL_OPENSSL FALSE)
    if(SYSTEM_OPENSSL_PREFIX)
        set(OPENSSL_ROOT_DIR ${SYSTEM_OPENSSL_PREFIX})
    endif()
endif()

#set ssl dynamic path
if (WIN32 AND MSVC)
    set(SSL_SUB_PATH "/bin/")
    if("${CMAKE_SIZEOF_VOID_P}" STREQUAL "8")
        set(SSL_DYNAMIC_LIB_NAME "libssl-${KMC_OPENSSL_VERSION}-x64.dll")
        set(SSL_DYNAMIC_LIB_LINK_NAME "libssl-${KMC_OPENSSL_VERSION}-x64.dll")
        set(CRYPTO_DYNAMIC_LIB_NAME "libcrypto-${KMC_OPENSSL_VERSION}-x64.dll")
        set(CRYPTO_DYNAMIC_LIB_LINK_NAME "libcrypto-${KMC_OPENSSL_VERSION}-x64.dll")
    else()
        set(SSL_DYNAMIC_LIB_NAME "libssl-${KMC_OPENSSL_VERSION}.dll")
        set(SSL_DYNAMIC_LIB_LINK_NAME "libssl-${KMC_OPENSSL_VERSION}.dll")
        set(CRYPTO_DYNAMIC_LIB_NAME "libcrypto-${KMC_OPENSSL_VERSION}.dll")
        set(CRYPTO_DYNAMIC_LIB_LINK_NAME "libcrypto-${KMC_OPENSSL_VERSION}.dll")
    endif()
elseif(UNIX)
    set(SSL_SUB_PATH "/lib/")
    set(SSL_DYNAMIC_LIB_NAME "libssl.so.${KMC_OPENSSL_VERSION}")
    set(SSL_DYNAMIC_LIB_LINK_NAME "libssl.so")
    set(CRYPTO_DYNAMIC_LIB_NAME "libcrypto.so.${KMC_OPENSSL_VERSION}")
    set(CRYPTO_DYNAMIC_LIB_LINK_NAME "libcrypto.so")
else()
    message(FATAL_ERROR "not supported platform")
endif()

if(LOCAL_OPENSSL)
    check_define(KMC_ALL_BINARY_DIR)
    set(OPENSSL_LOCAL_INSTALL_DIR "${KMC_ALL_BINARY_DIR}/openssl_root")

    file(TO_NATIVE_PATH "${OPENSSL_LOCAL_INSTALL_DIR}${SSL_SUB_PATH}${SSL_DYNAMIC_LIB_LINK_NAME}" SSL_IMPORTED_LOCATION_NAME)
    file(TO_NATIVE_PATH "${OPENSSL_LOCAL_INSTALL_DIR}${SSL_SUB_PATH}${CRYPTO_DYNAMIC_LIB_LINK_NAME}" CRYPTO_IMPORTED_LOCATION_NAME)

    set_target_properties(openssl_ssl_lib PROPERTIES
        IMPORTED_LOCATION "${SSL_IMPORTED_LOCATION_NAME}"
        PUBLIC_HEADER_DIR "${OPENSSL_LOCAL_INSTALL_DIR}/include"
        )

    set_target_properties(openssl_crypto_lib PROPERTIES
        IMPORTED_LOCATION "${CRYPTO_IMPORTED_LOCATION_NAME}"
        PUBLIC_HEADER_DIR "${OPENSSL_LOCAL_INSTALL_DIR}/include"
        )

    if(WIN32 AND MSVC)
        set_target_properties(openssl_ssl_lib PROPERTIES
            IMPORTED_IMPLIB "${OPENSSL_LOCAL_INSTALL_DIR}/lib/libssl.lib"
            )
        set_target_properties(openssl_crypto_lib PROPERTIES
            IMPORTED_IMPLIB "${OPENSSL_LOCAL_INSTALL_DIR}/lib/libcrypto.lib"
            )
    endif()

else()

    #minimum require openssl version
    find_package(OpenSSL ${OPENSSL_MINIMUM_REQUIRED_VERSION})
    if(NOT OPENSSL_FOUND)
        message(FATAL_ERROR "not find openssl libs pls install first")
    endif()

    check_define(OPENSSL_INCLUDE_DIR)
    check_define(OPENSSL_SSL_LIBRARY)
    check_define(OPENSSL_CRYPTO_LIBRARY)
    get_filename_component(OPENSSL_TOPLEVEL_DIR ${OPENSSL_INCLUDE_DIR} DIRECTORY)

    set_target_properties(openssl_ssl_lib PROPERTIES
        IMPORTED_LOCATION "${OPENSSL_SSL_LIBRARY}"
        PUBLIC_HEADER_DIR "${OPENSSL_INCLUDE_DIR}"
        )

    set_target_properties(openssl_crypto_lib PROPERTIES
        IMPORTED_LOCATION "${OPENSSL_CRYPTO_LIBRARY}"
        PUBLIC_HEADER_DIR "${OPENSSL_INCLUDE_DIR}"
        )
    if(WIN32 AND MSVC)
        #fix imported_location use dll
        file(GLOB TMP_CRYPTO_DLL "${OPENSSL_TOPLEVEL_DIR}/bin/${CRYPTO_DYNAMIC_LIB_NAME}")
        file(GLOB TMP_SSL_DLL "${OPENSSL_TOPLEVEL_DIR}/bin/${SSL_DYNAMIC_LIB_NAME}")
        if(NOT TMP_CRYPTO_DLL OR NOT TMP_SSL_DLL)
            message(FATAL_ERROR "Not find crypto dll or ssl dll")
        endif()
        set_target_properties(openssl_ssl_lib PROPERTIES
            IMPORTED_LOCATION "${TMP_SSL_DLL}"
            IMPORTED_IMPLIB "${OPENSSL_SSL_LIBRARY}"
            )
        set_target_properties(openssl_crypto_lib PROPERTIES
            IMPORTED_LOCATION "${TMP_CRYPTO_DLL}"
            IMPORTED_IMPLIB "${OPENSSL_CRYPTO_LIBRARY}"
            )
    endif()
endif()





