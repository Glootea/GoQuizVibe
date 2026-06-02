package com.glootea.goquiz.feature.auth.domain.usecase

import com.glootea.goquiz.feature.auth.domain.model.User
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import dev.zacsweers.metro.Inject

@Inject
class RegisterUseCase(
    private val repository: AuthRepository
) {
    suspend operator fun invoke(name: String, email: String, password: String): User {
        require(name.isNotBlank()) { "Name must not be blank" }
        require(email.isNotBlank()) { "Email must not be blank" }
        require(password.length >= MIN_PASSWORD_LENGTH) {
            "Password must be at least $MIN_PASSWORD_LENGTH characters"
        }
        return repository.register(name.trim(), email.trim(), password)
    }

    companion object {
        const val MIN_PASSWORD_LENGTH: Int = 6
    }
}
