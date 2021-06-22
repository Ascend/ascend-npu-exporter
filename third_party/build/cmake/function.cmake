
#functions 

function(check_cmd cmd)
    find_program(CMD_${cmd} ${cmd})
    if(NOT CMD_${cmd})
        message(FATAL_ERROR "${cmd} command not found")
    endif()
endfunction()


function(check_define var)
    if (NOT DEFINED ${var})
        message(FATAL_ERROR "missing ${var} defined check your config")
    endif()
endfunction()

function(get_sources dir pattern sources)
    file(GLOB_RECURSE ALLSRCS  ${dir}/${pattern})
    if(WIN32)
        file(GLOB_RECURSE TO_REMOVE_FILES ${dir}/*_unix.*)
        if(TO_REMOVE_FILES)
            list(REMOVE_ITEM ALLSRCS ${TO_REMOVE_FILES})
        endif()
    else()
        file(GLOB_RECURSE TO_REMOVE_FILES ${dir}/*_windows.*)
        if(TO_REMOVE_FILES)
            list(REMOVE_ITEM ALLSRCS ${TO_REMOVE_FILES})
        endif()
    endif()
    set(${sources} ${ALLSRCS} PARENT_SCOPE)
endfunction()


