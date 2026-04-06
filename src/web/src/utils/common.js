const validateForm = (formRef) => {
  return new Promise((resolve) => {
    formRef.validate((valid) => {
      resolve(valid)
    })
  })
}

export { validateForm }
