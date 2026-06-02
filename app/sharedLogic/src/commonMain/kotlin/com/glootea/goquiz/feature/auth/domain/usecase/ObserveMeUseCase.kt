package com.glootea.goquiz.feature.auth.domain.usecase

import com.glootea.goquiz.feature.auth.domain.model.User
import com.glootea.goquiz.feature.auth.domain.repository.AuthRepository
import dev.zacsweers.metro.Inject

@Inject
class ObserveMeUseCase(
    private val repository: AuthRepository
) {
    suspend operator fun invoke(): User? = repository.me()
}
