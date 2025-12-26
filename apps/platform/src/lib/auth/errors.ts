export class ApiError extends Error {
    status: number
    statusText: string
    body: string

    constructor(status: number, statusText: string, body: string) {
        super(`${status} ${statusText}: ${body}`)
        this.status = status
        this.statusText = statusText
        this.body = body
        this.name = 'ApiError'
    }

    isStatus(status: number): boolean {
        return this.status === status
    }
}
