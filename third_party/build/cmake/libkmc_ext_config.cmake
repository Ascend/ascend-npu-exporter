
check_define(KMCEXT_VERSION)
check_define(KMCEXT_C_SOURCE_DIR)
check_define(KMC_ALL_BINARY_DIR)
check_define(PROJECT_KMC_EXT)

if (NOT BUILD_STATIC)
    add_library(libkmc_ext_shared SHARED IMPORTED)
    set_target_properties(libkmc_ext_shared PROPERTIES
        PUBLIC_HEADER_DIR "${KMCEXT_C_SOURCE_DIR}/include")
else()
    add_library(libkmc_ext_shared STATIC IMPORTED)
    set_target_properties(libkmc_ext_shared PROPERTIES
        PUBLIC_HEADER_DIR "${KMCEXT_C_SOURCE_DIR}/include")
endif()

if(UNIX)
    set_target_properties(libkmc_ext_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/libkmcext.so.${KMCEXT_VERSION}"
        )
    if (BUILD_STATIC)
        set_target_properties(libkmc_ext_shared PROPERTIES
            IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/libkmcext.a"
            )
    endif()
elseif(WIN32 AND MSVC)
    set_target_properties(libkmc_ext_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_KMC_EXT}.dll"
        IMPORTED_IMPLIB "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_KMC_EXT}.lib"
        )
else()
    message(FATAL_ERROR "not supported platform")
endif()
