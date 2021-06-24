
include(${CMAKE_CURRENT_LIST_DIR}/function.cmake)

#common cmd for compile
if(NOT KMC_ONLY)
    find_package(Java)
endif()

if(UNIX)
    check_cmd(ln)
    check_cmd(ldconfig)
    check_cmd(cp)
    check_cmd(gzip)
    check_cmd(tar)
elseif(WIN32)
    check_cmd(cl)
    check_cmd(cmd)
    check_cmd(nmake)
    check_cmd(xcopy)
endif()

