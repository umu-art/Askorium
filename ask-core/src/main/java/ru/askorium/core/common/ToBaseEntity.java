package ru.askorium.core.common;

import org.mapstruct.Mapping;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;

@Retention(RetentionPolicy.CLASS)
@Mapping(target = "id", ignore = true)
@Mapping(target = "created", ignore = true)
@Mapping(target = "updated", ignore = true)
public @interface ToBaseEntity {
}
