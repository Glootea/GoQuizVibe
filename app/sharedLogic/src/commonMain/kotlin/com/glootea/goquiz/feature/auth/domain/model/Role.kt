package com.glootea.goquiz.feature.auth.domain.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class Role {
    @SerialName("teacher")
    Teacher,

    @SerialName("student")
    Student
}
