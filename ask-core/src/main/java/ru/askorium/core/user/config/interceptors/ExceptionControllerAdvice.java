package ru.askorium.core.user.config.interceptors;

import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.TypeMismatchException;
import org.springframework.http.HttpStatus;
import org.springframework.http.HttpStatusCode;
import org.springframework.http.ResponseEntity;
import org.springframework.http.converter.HttpMessageConversionException;
import org.springframework.validation.BindException;
import org.springframework.web.ErrorResponse;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;
import ru.askorium.core.exceptions.AskCoreException;

@Slf4j
@ControllerAdvice
public class ExceptionControllerAdvice {

    @ExceptionHandler
    public ResponseEntity<String> handle(Exception ex) {
        log.error("Ошибка при обработке запроса: ", ex);

        var res = parse(ex);

        return ResponseEntity
                .status(res.errorCode)
                .body(res.errorMessage);
    }

    private ErrorData parse(Exception ex) {
        if (ex instanceof AskCoreException askCoreException) {
            return new ErrorData(
                    HttpStatusCode.valueOf(askCoreException.getCode()),
                    ex.getLocalizedMessage());
        }

        if (ex instanceof BindException || ex instanceof HttpMessageConversionException || ex instanceof TypeMismatchException) {
            return new ErrorData(
                    HttpStatus.BAD_REQUEST,
                    ex.getLocalizedMessage());
        }

        if (ex instanceof ErrorResponse errorResponse) {
            return new ErrorData(
                    errorResponse.getStatusCode(),
                    errorResponse.getBody().getTitle());
        }

        return new ErrorData(
                HttpStatus.INTERNAL_SERVER_ERROR,
                "Internal server error");
    }

    record ErrorData(HttpStatusCode errorCode, String errorMessage) {
    }
}
