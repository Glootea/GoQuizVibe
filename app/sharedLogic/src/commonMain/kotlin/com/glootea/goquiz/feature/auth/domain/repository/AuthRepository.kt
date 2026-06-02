package com.glootea.goquiz.feature.auth.domain.repository

import com.glootea.goquiz.feature.auth.domain.model.User

interface AuthRepository {
    suspend fun login(email: String, password: String): User
    suspend fun register(name: String, email: String, password: String): User
    suspend fun logout()
    suspend fun me(): User?
}
