package com.glootea.goquiz.feature.auth.domain.usecase

import com.glootea.goquiz.feature.auth.domain.model.User
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import dev.zacsweers.metro.Inject

@Inject
class LoginUseCase(
    private val repository: AuthRepository
) {
    suspend operator fun invoke(email: String, password: String): User {
        require(email.isNotBlank()) { "Email must not be blank" }
        require(password.isNotEmpty()) { "Password must not be empty" }
        return repository.login(email.trim(), password)
    }
}
