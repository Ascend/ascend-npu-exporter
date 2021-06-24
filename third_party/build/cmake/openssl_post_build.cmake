
if(NOT EXISTS ${FROM_SO_NAME})
    message(FATAL_ERROR "file ${FROM_SO_NAME} not exists")
endif()

file(COPY ${FROM_SO_NAME} DESTINATION ${TO_DIR})

if(UNIX)
    file(GLOB SO_FILE ${FROM_SO_NAME}*${SO_VERSION})
    if(NOT SO_FILE)
        message(FATAL_ERROR "not find so links")
    endif()
    file(COPY ${SO_FILE} DESTINATION ${TO_DIR}
        FILE_PERMISSIONS OWNER_READ OWNER_EXECUTE)
endif()

