FROM python:3.14-slim
WORKDIR /application

COPY /ask-trafilatura/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY /api/build/python-parser-api/dist/python_parser_api-*-py3-none-any.whl .
COPY /api/build/python-renderer-api/dist/python_renderer_api-*-py3-none-any.whl .
RUN pip install --no-cache-dir ./*.whl && rm -f ./*.whl

COPY /ask-trafilatura/app/ app/

ENTRYPOINT ["python", "-m", "app.main"]