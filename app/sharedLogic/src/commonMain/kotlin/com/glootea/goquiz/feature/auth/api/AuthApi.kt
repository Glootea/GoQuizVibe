package com.glootea.goquiz.feature.auth.api

import com.glootea.goquiz.feature.auth.api.dto.AuthResponseDto
import com.glootea.goquiz.feature.auth.api.dto.LoginRequestDto
import com.glootea.goquiz.feature.auth.api.dto.MeResponseDto
import com.glootea.goquiz.feature.auth.api.dto.RegisterRequestDto
import com.glootea.goquiz.feature.auth.domain.model.User

interface AuthApi {
    suspend fun login(request: LoginRequestDto): AuthResponseDto
    suspend fun register(request: RegisterRequestDto): AuthResponseDto
    suspend fun logout()
    suspend fun me(): MeResponseDto?
}
