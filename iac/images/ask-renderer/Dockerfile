FROM mcr.microsoft.com/playwright/python:v1.58.0-noble
WORKDIR /application


COPY /ask-renderer/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY /api/build/python-renderer-api/dist/python_renderer_api-*-py3-none-any.whl .
RUN pip install --no-cache-dir ./*.whl && rm -f ./*.whl

COPY /ask-renderer/app/ app/

EXPOSE 8080

ENTRYPOINT ["python", "-m", "app.main"]