package com.glootea.goquiz.feature.auth.domain.usecase

import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import dev.zacsweers.metro.Inject

@Inject
class LogoutUseCase(
    private val repository: AuthRepository
) {
    suspend operator fun invoke() {
        repository.logout()
    }
}
