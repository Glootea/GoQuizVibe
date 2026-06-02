package com.glootea.goquiz.core.http

import kotlinx.serialization.Serializable

@Serializable
data class ErrorBodyDto(
    val code: String,
    val message: String,
    val fields: Map<String, String> = emptyMap()
)

@Serializable
data class ErrorDto(
    val error: ErrorBodyDto
)
