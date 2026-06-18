package com.glootea.goquiz.core.error

sealed class AppError : RuntimeException() {
    class Unauthorized : AppError()
    class Forbidden : AppError()
    data class NotFound(val resource: String) : AppError()
    data class Conflict(val code: String) : AppError()
    data class Validation(val fields: Map<String, String>) : AppError()
    data class Network(override val cause: Throwable) : AppError()
    data class Unknown(override val cause: Throwable) : AppError()
}
