package com.glootea.goquiz

interface Platform {
    val name: String
}

expect fun getPlatform(): Platform