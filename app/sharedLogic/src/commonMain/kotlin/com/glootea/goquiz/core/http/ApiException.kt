package com.glootea.goquiz.core.http

class ApiException(
    val code: String,
    override val message: String
) : RuntimeException(message)


