package com.glootea.goquiz.feature.auth.api.dto

import kotlinx.serialization.Serializable

@Serializable
data class LoginRequestDto(
    val email: String,
    val password: String
)

@Serializable
data class RegisterRequestDto(
    val name: String,
    val email: String,
    val password: String
)

@Serializable
data class UserDto(
    val id: String,
    val name: String,
    val email: String,
    val role: String
)

@Serializable
data class AuthResponseDto(
    val user: UserDto,
    val access_token: String,
    val expires_at: String
)

@Serializable
data class MeResponseDto(
    val user: UserDto
)
