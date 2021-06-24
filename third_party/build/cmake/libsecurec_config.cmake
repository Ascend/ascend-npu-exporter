#hwsec c

include(${CMAKE_CURRENT_LIST_DIR}/function.cmake)
include(${CMAKE_CURRENT_LIST_DIR}/path_config.cmake)

check_define(PROJECT_HWSEC)
check_define(KMC_ALL_BINARY_DIR)



if (NOT BUILD_STATIC)
    add_library(libsecurec_shared SHARED IMPORTED)
    set_target_properties(libsecurec_shared PROPERTIES
        IMPORTED_NO_SONAME true
        PUBLIC_HEADER_DIR "${KMC_EXTERNAL_HWSEC_C_SOURCE_DIR}/include"
    )
else()
    add_library(libsecurec_shared STATIC IMPORTED)
    set_target_properties(libsecurec_shared PROPERTIES
        IMPORTED_NO_SONAME true
        PUBLIC_HEADER_DIR "${KMC_EXTERNAL_HWSEC_C_SOURCE_DIR}/include"
    )
endif()

if(UNIX)
    set_target_properties(libsecurec_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/lib${PROJECT_HWSEC}.so"
        )
    if (BUILD_STATIC)
        set_target_properties(libsecurec_shared PROPERTIES
            IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/lib${PROJECT_HWSEC}.a"
        )
    endif()
elseif(WIN32 AND MSVC)
    set_target_properties(libsecurec_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_HWSEC}.dll"
        IMPORTED_IMPLIB "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_HWSEC}.lib"
        )
else()
    message(FATAL_ERROR "not supported platform")
endif()
