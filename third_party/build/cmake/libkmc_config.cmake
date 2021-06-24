#kmc sdp 默认编译时没有soname 需要特殊处理

check_define(KMC_SDP_VERSION)
check_define(KMC_C_SOURCE_DIR)
check_define(KMC_ALL_BINARY_DIR)
check_define(PROJECT_KMC)
check_define(PROJECT_SDP)

# KMC 加解密相关接口
if (NOT BUILD_STATIC)
    add_library(libkmc_shared SHARED IMPORTED)
    set_target_properties(libkmc_shared PROPERTIES 
        PUBLIC_HEADER_DIR "${KMC_C_SOURCE_DIR}/include")

    # SDP 敏感数据加密相关接口
    add_library(libsdp_shared SHARED IMPORTED)
    set_target_properties(libsdp_shared PROPERTIES
        PUBLIC_HEADER_DIR "${KMC_C_SOURCE_DIR}/src/sdp")
else()
    add_library(libkmc_shared STATIC IMPORTED)
    set_target_properties(libkmc_shared PROPERTIES 
        PUBLIC_HEADER_DIR "${KMC_C_SOURCE_DIR}/include")

    # SDP 敏感数据加密相关接口
    add_library(libsdp_shared STATIC IMPORTED)
    set_target_properties(libsdp_shared PROPERTIES
        PUBLIC_HEADER_DIR "${KMC_C_SOURCE_DIR}/src/sdp")
endif()

if(UNIX)
    set_target_properties(libkmc_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/lib${PROJECT_KMC}.so.${KMC_SDP_VERSION}"
        )
    set_target_properties(libsdp_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/lib${PROJECT_SDP}.so.${KMC_SDP_VERSION}"
        )
    if (BUILD_STATIC)
        set_target_properties(libkmc_shared PROPERTIES
            IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/lib${PROJECT_KMC}.a"
            )
        set_target_properties(libsdp_shared PROPERTIES
            IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/lib${PROJECT_SDP}.a"
            )
    endif()
elseif(WIN32 AND MSVC)
    set_target_properties(libkmc_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_KMC}.dll"
        IMPORTED_IMPLIB "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_KMC}.lib"
        )
    set_target_properties(libsdp_shared PROPERTIES
        IMPORTED_LOCATION "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_SDP}.dll"
        IMPORTED_IMPLIB "${KMC_ALL_BINARY_DIR}/lib/${PROJECT_SDP}.lib"
        )
else()
    message(FATAL_ERROR "not supported platform")
endif()
