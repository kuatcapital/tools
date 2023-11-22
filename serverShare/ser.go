package main

import (
        "net/http"
)

func main() {
        fs := http.FileServer(http.Dir("директория которую хотим расшарить"))

        http.Handle("/адрес/", http.StripPrefix("/адрес/", fs))

        // Запустите HTTP-сервер на порту 8080.
        if err := http.ListenAndServe(":8080", nil); err != nil {
                panic(err)
        }
}
