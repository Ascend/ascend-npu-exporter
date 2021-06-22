#securec sub cmake

check_define(PROJECT_HWSEC)
check_define(KMC_EXTERNAL_HWSEC_C_SOURCE_DIR)

project ("${PROJECT_HWSEC}")

include_directories(${KMC_EXTERNAL_HWSEC_C_SOURCE_DIR}/include)

message(${KMC_EXTERNAL_HWSEC_C_SOURCE_DIR})

set(HW_SEC_LIBSRCS ${KMC_EXTERNAL_HWSEC_C_SOURCE_DIR}/src)
file(GLOB_RECURSE HWSEC_SRCS ${HW_SEC_LIBSRCS}/*.c)
file(GLOB_RECURSE HWSEC_HDRS ${HW_SEC_LIBSRCS}/include/*.h)

include(${KMC_ALL_TOPLEVEL_CMAKE_DIR}/cmake/security_compile.cmake)


if(WIN32 AND MSVC)
    foreach(ITEM IN LISTS HWSEC_SRCS)
        string(FIND ${ITEM} "memset_s" IS_FIND_MEMSET_S)
        if(NOT IS_FIND_MEMSET_S EQUAL -1)
            set(HWSEC_FILTED_SRC ${HWSEC_FILTED_SRC} ${ITEM})
        endif()
    endforeach()
elseif(UNIX)
    set(HWSEC_FILTED_SRC ${HWSEC_SRCS})
else()
    message(FATAL_ERROR "Not Supported")
endif()

add_library("${PROJECT_HWSEC}" ${BUILD_LIB_TYPE} ${HWSEC_FILTED_SRC} ${HWSEC_HDRS})

if(WIN32 AND MSVC)
    target_compile_options(${PROJECT_HWSEC}
        #windows has own security functions
        PUBLIC -DSECUREC_ONLY_DECLARE_MEMSET
        PUBLIC -DSECUREC_IS_DLL_LIBRARY
        PUBLIC -DSECUREC_DLL_EXPORT
        )
endif()

set_target_properties("${PROJECT_HWSEC}" PROPERTIES 
    #VERSION ${HWSEC_VERSION}
    #disable SONAME  
    NO_SONAME TRUE
)
