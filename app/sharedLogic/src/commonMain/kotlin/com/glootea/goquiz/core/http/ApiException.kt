package com.glootea.goquiz.core.http

import com.glootea.goquiz.core.error.AppError
import io.ktor.client.plugins.ClientRequestException
import io.ktor.client.plugins.HttpRequestTimeoutException
import io.ktor.client.plugins.RedirectResponseException
import io.ktor.client.plugins.ResponseException
import io.ktor.client.plugins.ServerResponseException
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess
import io.ktor.serialization.JsonConvertException
import kotlinx.serialization.json.Json

class ApiException(
    val status: HttpStatusCode,
    val code: String,
    override val message: String,
    val fields: Map<String, String> = emptyMap()
) : RuntimeException(message)

private val errorJson = Json { ignoreUnknownKeys = true; isLenient = true }

suspend fun HttpResponse.parseErrorOrThrow() {
    if (status.isSuccess()) return
    val raw = runCatching { bodyAsText() }.getOrDefault("")
    val parsed: ErrorDto? = runCatching { errorJson.decodeFromString(ErrorDto.serializer(), raw) }.getOrNull()
    val body = parsed?.error
    throw ApiException(
        status = status,
        code = body?.code ?: statusCodeToCode(status),
        message = body?.message ?: "Request failed with status ${status.value}",
        fields = body?.fields.orEmpty()
    )
}

fun throwApiException(
    response: HttpResponse,
    cause: Throwable
): Nothing {
    if (response.status.isSuccess()) {
        throw IllegalStateException("throwApiException called on success response", cause)
    }
    throw mapToAppError(cause, response).asApiException(cause)
}

private fun mapToAppError(cause: Throwable, response: HttpResponse): AppError = when (cause) {
    is ApiException -> when (cause.status.value) {
        401 -> AppError.Unauthorized
        403 -> AppError.Forbidden
        404 -> AppError.NotFound("")
        409 -> AppError.Conflict(cause.code.ifBlank { "CONFLICT" })
        in 400..499 -> AppError.Validation(cause.fields)
        else -> AppError.Unknown(cause)
    }
    is HttpRequestTimeoutException -> AppError.Network(cause)
    is RedirectResponseException,
    is ClientRequestException,
    is ServerResponseException,
    is ResponseException -> when (response.status.value) {
        401 -> AppError.Unauthorized
        403 -> AppError.Forbidden
        404 -> AppError.NotFound("")
        409 -> AppError.Conflict("CONFLICT")
        in 400..499 -> AppError.Validation(emptyMap())
        else -> AppError.Unknown(cause)
    }
    is JsonConvertException -> AppError.Network(cause)
    else -> AppError.Network(cause)
}

private fun AppError.asApiException(cause: Throwable): Nothing {
    when (this) {
        AppError.Unauthorized -> throw ApiException(HttpStatusCode.Unauthorized, "UNAUTHORIZED", "Unauthorized")
        AppError.Forbidden -> throw ApiException(HttpStatusCode.Forbidden, "FORBIDDEN", "Forbidden")
        is AppError.NotFound -> throw ApiException(HttpStatusCode.NotFound, "NOT_FOUND", resource.ifBlank { "Not found" })
        is AppError.Conflict -> throw ApiException(HttpStatusCode.Conflict, code, "Conflict")
        is AppError.Validation -> throw ApiException(HttpStatusCode.BadRequest, "VALIDATION", "Validation failed", fields)
        is AppError.Network -> throw ApiException(HttpStatusCode.ServiceUnavailable, "NETWORK", cause.message ?: "Network error")
        is AppError.Unknown -> throw ApiException(HttpStatusCode.InternalServerError, "INTERNAL", cause.message ?: "Internal error")
    }
}

private fun statusCodeToCode(status: HttpStatusCode): String = when (status.value) {
    400 -> "BAD_REQUEST"
    401 -> "UNAUTHORIZED"
    403 -> "FORBIDDEN"
    404 -> "NOT_FOUND"
    409 -> "CONFLICT"
    422 -> "UNPROCESSABLE_ENTITY"
    429 -> "TOO_MANY_REQUESTS"
    in 500..599 -> "INTERNAL"
    else -> "ERROR"
}
